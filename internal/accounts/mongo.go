// Package accounts stores users, login sessions and each user's progress and
// code drafts in MongoDB. It is used only when the server is started with a
// MongoDB URI; without one the app stays single-user on plain files.
package accounts

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"lldprep/internal/store"
)

const SessionTTL = 30 * 24 * time.Hour

var (
	ErrUserExists      = errors.New("that username is taken")
	ErrInvalidUsername = errors.New("username must be 3–32 characters: letters, digits, '.', '_' or '-'")
	ErrWeakPassword    = errors.New("password must be at least 8 characters")
)

var usernameRe = regexp.MustCompile(`^[a-z0-9._-]{3,32}$`)

// NormalizeUsername lower-cases and trims a username so logins are case-insensitive.
func NormalizeUsername(u string) string { return strings.ToLower(strings.TrimSpace(u)) }

// Mongo implements store.Store with one progress document and one document
// per draft for every user.
type Mongo struct {
	client                                        *mongo.Client
	users, sessions, progress, drafts, identities *mongo.Collection
	dummyHash                                     []byte
}

var _ store.Store = (*Mongo)(nil)

// Connect opens the database, checks it is reachable and ensures indexes.
func Connect(ctx context.Context, uri, dbName string) (*Mongo, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetAppName("lldprep"))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(context.Background())
		return nil, err
	}
	db := client.Database(dbName)
	m := &Mongo{
		client:     client,
		users:      db.Collection("users"),
		sessions:   db.Collection("sessions"),
		progress:   db.Collection("progress"),
		drafts:     db.Collection("drafts"),
		identities: db.Collection("identities"),
	}
	// Expired sessions are removed by MongoDB itself.
	if _, err := m.sessions.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expiresAt", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}); err != nil {
		return nil, err
	}
	if _, err := m.drafts.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "user", Value: 1}}}); err != nil {
		return nil, err
	}
	if _, err := m.identities.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "user", Value: 1}}}); err != nil {
		return nil, err
	}
	// Compared against when a username doesn't exist, so a failed login takes
	// as long whether or not the account is real.
	m.dummyHash, _ = bcrypt.GenerateFromPassword([]byte("not-a-real-password"), bcrypt.DefaultCost)
	return m, nil
}

func (m *Mongo) Close() error { return m.client.Disconnect(context.Background()) }

// ---- users ----

type userDoc struct {
	ID           string    `bson:"_id"`
	PasswordHash []byte    `bson:"passwordHash,omitempty"` // empty for Google/Apple-only users
	CreatedAt    time.Time `bson:"createdAt"`
}

// CreateUser registers a new user. The username must already be normalized.
func (m *Mongo) CreateUser(ctx context.Context, username, password string) error {
	if !usernameRe.MatchString(username) {
		return ErrInvalidUsername
	}
	if len(password) < 8 {
		return ErrWeakPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = m.users.InsertOne(ctx, userDoc{ID: username, PasswordHash: hash, CreatedAt: time.Now().UTC()})
	if mongo.IsDuplicateKeyError(err) {
		return ErrUserExists
	}
	return err
}

// CheckPassword reports whether the username exists and the password matches.
func (m *Mongo) CheckPassword(ctx context.Context, username, password string) (bool, error) {
	var u userDoc
	err := m.users.FindOne(ctx, bson.D{{Key: "_id", Value: username}}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		bcrypt.CompareHashAndPassword(m.dummyHash, []byte(password))
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if len(u.PasswordHash) == 0 { // signed up with Google/Apple: no password to match
		bcrypt.CompareHashAndPassword(m.dummyHash, []byte(password))
		return false, nil
	}
	return bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password)) == nil, nil
}

// HasPassword reports whether the user can sign in with a password.
func (m *Mongo) HasPassword(ctx context.Context, username string) (bool, error) {
	var u userDoc
	err := m.users.FindOne(ctx, bson.D{{Key: "_id", Value: username}}).Decode(&u)
	if err != nil {
		return false, err
	}
	return len(u.PasswordHash) > 0, nil
}

// ---- external identities (Google, Apple) ----

var ErrIdentityLinked = errors.New("that account is already connected to another user")

type identityDoc struct {
	ID        string    `bson:"_id"` // provider + ":" + subject
	Provider  string    `bson:"provider"`
	Subject   string    `bson:"subject"`
	User      string    `bson:"user"`
	Email     string    `bson:"email,omitempty"`
	CreatedAt time.Time `bson:"createdAt"`
}

func identityID(provider, subject string) string { return provider + ":" + subject }

// FindIdentity returns the user an external identity is linked to.
func (m *Mongo) FindIdentity(ctx context.Context, provider, subject string) (string, bool, error) {
	var d identityDoc
	err := m.identities.FindOne(ctx, bson.D{{Key: "_id", Value: identityID(provider, subject)}}).Decode(&d)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return d.User, true, nil
}

// LinkIdentity connects an external identity to a user. An identity belongs to
// at most one user; linking it again to the same user is a no-op.
func (m *Mongo) LinkIdentity(ctx context.Context, provider, subject, user, email string) error {
	_, err := m.identities.InsertOne(ctx, identityDoc{
		ID: identityID(provider, subject), Provider: provider, Subject: subject,
		User: user, Email: email, CreatedAt: time.Now().UTC(),
	})
	if mongo.IsDuplicateKeyError(err) {
		owner, _, ferr := m.FindIdentity(ctx, provider, subject)
		if ferr == nil && owner == user {
			return nil
		}
		return ErrIdentityLinked
	}
	return err
}

// Providers lists the external providers linked to a user, sorted.
func (m *Mongo) Providers(ctx context.Context, user string) ([]string, error) {
	cur, err := m.identities.Find(ctx, bson.D{{Key: "user", Value: user}}, options.Find().SetSort(bson.D{{Key: "provider", Value: 1}}))
	if err != nil {
		return nil, err
	}
	var docs []identityDoc
	if err := cur.All(ctx, &docs); err != nil {
		return nil, err
	}
	out := []string{}
	for _, d := range docs {
		if len(out) == 0 || out[len(out)-1] != d.Provider {
			out = append(out, d.Provider)
		}
	}
	return out, nil
}

var notUsernameChars = regexp.MustCompile(`[^a-z0-9._-]+`)

// CreateExternalUser creates a password-less user for a Google/Apple sign-up,
// with a username derived from `hint` (usually the email's local part) and a
// numeric suffix if that's taken. It returns the username.
func (m *Mongo) CreateExternalUser(ctx context.Context, hint string) (string, error) {
	base := strings.Trim(notUsernameChars.ReplaceAllString(strings.ToLower(hint), "."), "._-")
	if len(base) > 28 {
		base = base[:28]
	}
	if len(base) < 3 {
		base += "user" // "" -> "user", "jo" -> "jouser"
	}
	for i := 1; i <= 50; i++ {
		name := base
		if i > 1 {
			name = fmt.Sprintf("%s%d", base, i)
		}
		_, err := m.users.InsertOne(ctx, userDoc{ID: name, CreatedAt: time.Now().UTC()})
		if err == nil {
			return name, nil
		}
		if !mongo.IsDuplicateKeyError(err) {
			return "", err
		}
	}
	return "", errors.New("could not find a free username")
}

// ---- sessions ----

type sessionDoc struct {
	ID        string    `bson:"_id"` // sha256 of the cookie token; the token itself is never stored
	User      string    `bson:"user"`
	ExpiresAt time.Time `bson:"expiresAt"`
}

func tokenID(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// NewSession creates a login session and returns its secret token.
func (m *Mongo) NewSession(ctx context.Context, user string) (string, time.Time, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", time.Time{}, err
	}
	token := hex.EncodeToString(b)
	exp := time.Now().UTC().Add(SessionTTL)
	_, err := m.sessions.InsertOne(ctx, sessionDoc{ID: tokenID(token), User: user, ExpiresAt: exp})
	return token, exp, err
}

// SessionUser returns the user a token belongs to, if the session is live.
func (m *Mongo) SessionUser(ctx context.Context, token string) (string, bool, error) {
	var s sessionDoc
	err := m.sessions.FindOne(ctx, bson.D{{Key: "_id", Value: tokenID(token)}}).Decode(&s)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	// The TTL monitor runs about once a minute, so check expiry here too.
	if time.Now().After(s.ExpiresAt) {
		return "", false, nil
	}
	return s.User, true, nil
}

func (m *Mongo) DeleteSession(ctx context.Context, token string) error {
	_, err := m.sessions.DeleteOne(ctx, bson.D{{Key: "_id", Value: tokenID(token)}})
	return err
}

// ---- progress (store.Store) ----

// The state is stored as a real BSON document, so it can be read and
// queried in MongoDB tools, not as an opaque string.
type progressDoc struct {
	ID        string    `bson:"_id"`
	State     bson.D    `bson:"state"`
	UpdatedAt time.Time `bson:"updatedAt"`
}

func (m *Mongo) LoadState(ctx context.Context, user string) (json.RawMessage, error) {
	var p progressDoc
	err := m.progress.FindOne(ctx, bson.D{{Key: "_id", Value: user}}).Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return json.RawMessage("{}"), nil
	}
	if err != nil {
		return nil, err
	}
	if p.State == nil {
		return json.RawMessage("{}"), nil
	}
	return bson.MarshalExtJSON(p.State, false, false)
}

func (m *Mongo) SaveState(ctx context.Context, user string, b []byte) error {
	if !store.ValidState(b) {
		return store.ErrInvalidState
	}
	var doc bson.D
	if err := bson.UnmarshalExtJSON(b, false, &doc); err != nil {
		return store.ErrInvalidState
	}
	_, err := m.progress.ReplaceOne(ctx, bson.D{{Key: "_id", Value: user}},
		progressDoc{ID: user, State: doc, UpdatedAt: time.Now().UTC()},
		options.Replace().SetUpsert(true))
	return err
}

// ---- drafts (store.Store) ----

type draftDoc struct {
	ID        string    `bson:"_id"` // user + "/" + key; neither may contain '/'
	User      string    `bson:"user"`
	Key       string    `bson:"key"`
	Code      string    `bson:"code"`
	UpdatedAt time.Time `bson:"updatedAt"`
}

func draftID(user, key string) (string, error) {
	if !store.ValidKey(key) {
		return "", store.ErrInvalidKey
	}
	return user + "/" + key, nil
}

func (m *Mongo) LoadDraft(ctx context.Context, user, key string) (string, bool, error) {
	id, err := draftID(user, key)
	if err != nil {
		return "", false, err
	}
	var d draftDoc
	err = m.drafts.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&d)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return d.Code, true, nil
}

func (m *Mongo) SaveDraft(ctx context.Context, user, key, code string) error {
	id, err := draftID(user, key)
	if err != nil {
		return err
	}
	_, err = m.drafts.ReplaceOne(ctx, bson.D{{Key: "_id", Value: id}},
		draftDoc{ID: id, User: user, Key: key, Code: code, UpdatedAt: time.Now().UTC()},
		options.Replace().SetUpsert(true))
	return err
}

func (m *Mongo) DeleteDraft(ctx context.Context, user, key string) error {
	id, err := draftID(user, key)
	if err != nil {
		return err
	}
	_, err = m.drafts.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	return err
}

// DropForTests deletes the whole database. Only tests call it.
func (m *Mongo) DropForTests(ctx context.Context) error {
	return m.users.Database().Drop(ctx)
}
