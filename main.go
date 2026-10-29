// Command lldprep serves the LLD interview prep app.
//
//	go run .                 # serve on http://127.0.0.1:8080 (single user, progress in data/)
//	go run . -verify         # check every reference solution passes its tests
//	MONGODB_URI=... go run . # accounts mode: sign-in, per-user progress in MongoDB
package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"lldprep/internal/accounts"
	"lldprep/internal/content"
	"lldprep/internal/runner"
	"lldprep/internal/store"
)

//go:embed web content
var embedded embed.FS

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "listen address")
	dataDir := flag.String("data", "data", "directory for progress and drafts")
	dev := flag.Bool("dev", false, "serve web/ and content/ from disk (live edits)")
	verify := flag.Bool("verify", false, "run every reference solution against its tests and exit")
	only := flag.String("only", "", "with -verify: only units whose id contains this")
	onlyLang := flag.String("lang", "", "with -verify: only this language (java, go or python)")
	noOpen := flag.Bool("no-open", false, "do not open the browser")
	password := flag.String("password", os.Getenv("LLDPREP_PASSWORD"), "require this password (HTTP basic auth); defaults to $LLDPREP_PASSWORD")
	mongoURI := flag.String("mongo", os.Getenv("MONGODB_URI"), "MongoDB URI: enables sign-in and per-user progress; defaults to $MONGODB_URI")
	mongoDB := flag.String("mongo-db", envOr("MONGODB_DB", "lldprep"), "MongoDB database name; defaults to $MONGODB_DB or lldprep")
	invite := flag.String("invite", os.Getenv("LLDPREP_INVITE_CODE"), "with -mongo: code required to sign up; defaults to $LLDPREP_INVITE_CODE")
	publicURL := flag.String("public-url", os.Getenv("LLDPREP_PUBLIC_URL"), "public address used in link previews, e.g. https://lld-prep.example.com; defaults to $LLDPREP_PUBLIC_URL, else the request's host")
	flag.Parse()

	// Hosting platforms such as Render pass the port in $PORT and expect the
	// app to listen on all interfaces; there is no browser to open there.
	if port := os.Getenv("PORT"); port != "" && !flagSet("addr") {
		*addr = "0.0.0.0:" + port
		*noOpen = true
	}

	webFS, contentFS := subFS(embedded, "web"), subFS(embedded, "content")
	if *dev {
		webFS, contentFS = os.DirFS("web"), os.DirFS("content")
	}
	cs, err := content.Load(contentFS)
	if err != nil {
		log.Fatalf("loading content: %v", err)
	}
	if *verify {
		os.Exit(runVerify(cs, *only, *onlyLang))
	}

	au := &auth{inviteCode: *invite, limiter: newLoginLimiter()}
	var st store.Store
	if *mongoURI != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		db, err := accounts.Connect(ctx, *mongoURI, *mongoDB)
		cancel()
		if err != nil {
			log.Fatalf("connecting to MongoDB: %v", err)
		}
		defer db.Close()
		au.db, st = db, db
		log.Printf("accounts mode: users and progress in MongoDB database %q", *mongoDB)
		if au.social, err = socialFromEnv(strings.TrimRight(*publicURL, "/")); err != nil {
			log.Fatalf("configuring Google/GitHub sign-in: %v", err)
		}
		if names := au.social.names(); len(names) > 0 {
			log.Printf("social sign-in: %s (callbacks at <public url>/api/auth/<provider>/callback)", strings.Join(names, ", "))
			if *publicURL == "" {
				log.Printf("WARNING: set -public-url / $LLDPREP_PUBLIC_URL so sign-in redirect URIs match what you registered with the provider.")
			}
		}
	} else {
		if os.Getenv("GOOGLE_CLIENT_ID") != "" || os.Getenv("GITHUB_CLIENT_ID") != "" {
			log.Printf("WARNING: Google/GitHub sign-in needs accounts mode; set -mongo / $MONGODB_URI. Ignoring the provider settings.")
		}
		files, err := store.OpenFiles(*dataDir)
		if err != nil {
			log.Fatalf("opening data dir: %v", err)
		}
		st = files
	}

	a := &app{content: cs, store: st, dev: *dev, contentFS: contentFS}
	mux := http.NewServeMux()
	a.routes(mux)
	au.routes(mux)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "ok") })
	mux.Handle("/", staticHandler(webFS, strings.TrimRight(*publicURL, "/")))
	handler := au.requireUser(mux)
	public := !isLoopback(*addr)
	if *password != "" {
		handler = basicAuth(*password, handler)
	} else if public && !au.enabled() {
		log.Printf("WARNING: listening on %s without a password. Anyone who can reach it can run code on this machine. Set -password or $LLDPREP_PASSWORD.", *addr)
	}
	if public && au.enabled() && *invite == "" {
		log.Printf("WARNING: anyone can sign up, and every account can run code on this machine. Set -invite or $LLDPREP_INVITE_CODE to restrict sign-ups.")
	}

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	// Open the exact address we bound: "localhost" may resolve to ::1, where
	// another program could be listening on the same port.
	url := "http://" + ln.Addr().String()
	tc := runner.Toolchains()
	log.Printf("LLD Prep running at %s", url)
	log.Printf("toolchains: java=%q go=%q python=%q", tc["java"], tc["go"], tc["python"])
	if !*noOpen {
		go openBrowser(url)
	}
	log.Fatal(http.Serve(ln, handler))
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func flagSet(name string) bool {
	set := false
	flag.Visit(func(f *flag.Flag) { set = set || f.Name == name })
	return set
}

func isLoopback(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return host == "localhost" || (ip != nil && ip.IsLoopback())
}

// basicAuth guards everything except /healthz with a single shared password;
// any username is accepted.
func basicAuth(password string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" || previewAssets[r.URL.Path] {
			h.ServeHTTP(w, r)
			return
		}
		_, got, ok := r.BasicAuth()
		if !ok || !equalSecret(got, password) {
			w.Header().Set("WWW-Authenticate", `Basic realm="LLD Prep", charset="UTF-8"`)
			http.Error(w, "password required", http.StatusUnauthorized)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// previewAssets are served without the -password prompt so that link-preview
// crawlers (WhatsApp, Slack, LinkedIn...) can fetch the image and icons.
var previewAssets = map[string]bool{"/og-image.png": true, "/favicon.svg": true, "/apple-touch-icon.png": true}

// staticHandler serves web/. Browsers may cache the vendored libraries but
// always revalidate the app's own files, so edits and upgrades show up
// immediately. index.html gets its __ORIGIN__ placeholders filled in, because
// link-preview tags need absolute URLs.
func staticHandler(webFS fs.FS, publicURL string) http.Handler {
	files := http.FileServer(http.FS(webFS))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			page, err := fs.ReadFile(webFS, "index.html")
			if err != nil {
				httpError(w, err)
				return
			}
			origin := publicURL
			if origin == "" {
				scheme := "http"
				if isHTTPS(r) {
					scheme = "https"
				}
				origin = scheme + "://" + r.Host
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-cache")
			w.Write(bytes.ReplaceAll(page, []byte("__ORIGIN__"), []byte(html.EscapeString(origin))))
			return
		}
		switch {
		case strings.HasPrefix(r.URL.Path, "/vendor/"):
			w.Header().Set("Cache-Control", "public, max-age=86400")
		case previewAssets[r.URL.Path]:
			w.Header().Set("Cache-Control", "public, max-age=3600")
		default:
			w.Header().Set("Cache-Control", "no-cache")
		}
		files.ServeHTTP(w, r)
	})
}

func subFS(f fs.FS, dir string) fs.FS {
	s, err := fs.Sub(f, dir)
	if err != nil {
		panic(err)
	}
	return s
}

func openBrowser(url string) {
	time.Sleep(300 * time.Millisecond)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

type app struct {
	mu        sync.RWMutex
	content   *content.Store
	contentFS fs.FS
	dev       bool
	store     store.Store
	runSlots  chan struct{}
}

// cs returns the content store, reloading from disk in -dev mode.
func (a *app) cs() *content.Store {
	if a.dev {
		if c, err := content.Load(a.contentFS); err == nil {
			a.mu.Lock()
			a.content = c
			a.mu.Unlock()
		} else {
			log.Printf("content reload failed: %v", err)
		}
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.content
}

func (a *app) routes(mux *http.ServeMux) {
	a.runSlots = make(chan struct{}, 2)
	mux.HandleFunc("GET /api/plan", func(w http.ResponseWriter, r *http.Request) {
		cs := a.cs()
		type mockOut struct {
			*content.Mock
			Difficulty string `json:"difficulty"`
		}
		mocks := []mockOut{}
		for _, m := range cs.Mocks {
			mocks = append(mocks, mockOut{m, cs.Exercises[m.Exercise].Difficulty})
		}
		writeJSON(w, map[string]any{"days": cs.Plan.Days, "mocks": mocks, "rubric": cs.Rubric, "catalog": catalog(cs)})
	})
	mux.HandleFunc("GET /api/lesson/{id}", func(w http.ResponseWriter, r *http.Request) {
		if l, ok := a.cs().Lessons[r.PathValue("id")]; ok {
			writeJSON(w, l)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("GET /api/exercise/{id}", func(w http.ResponseWriter, r *http.Request) {
		if e, ok := a.cs().Exercises[r.PathValue("id")]; ok {
			writeJSON(w, e)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("GET /api/tutorial/{id}", func(w http.ResponseWriter, r *http.Request) {
		if t, ok := a.cs().Tutorials[r.PathValue("id")]; ok {
			writeJSON(w, t)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("GET /api/quiz/{id}", func(w http.ResponseWriter, r *http.Request) {
		if q, ok := a.cs().Quizzes[r.PathValue("id")]; ok {
			writeJSON(w, q)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("GET /api/solution/{kind}/{id}/{lang}", func(w http.ResponseWriter, r *http.Request) {
		cu, ok := a.codeUnit(r.PathValue("kind"), r.PathValue("id"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, map[string]string{"code": cu.Reference[r.PathValue("lang")]})
	})
	mux.HandleFunc("GET /api/search", func(w http.ResponseWriter, r *http.Request) {
		hits := a.cs().Search(r.URL.Query().Get("q"))
		if hits == nil {
			hits = []content.Hit{}
		}
		writeJSON(w, hits)
	})
	mux.HandleFunc("GET /api/toolchains", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, runner.Toolchains())
	})
	mux.HandleFunc("GET /api/state", func(w http.ResponseWriter, r *http.Request) {
		b, err := a.store.LoadState(r.Context(), userOf(r))
		if err != nil {
			httpError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	})
	mux.HandleFunc("PUT /api/state", func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(io.LimitReader(r.Body, 4<<20))
		if err == nil {
			err = a.store.SaveState(r.Context(), userOf(r), b)
		}
		if errors.Is(err, store.ErrInvalidState) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err != nil {
			httpError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /api/draft/{key}", func(w http.ResponseWriter, r *http.Request) {
		code, ok, err := a.store.LoadDraft(r.Context(), userOf(r), r.PathValue("key"))
		if errors.Is(err, store.ErrInvalidKey) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err != nil {
			httpError(w, err)
			return
		}
		writeJSON(w, map[string]any{"code": code, "exists": ok})
	})
	mux.HandleFunc("PUT /api/draft/{key}", func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err == nil {
			err = a.store.SaveDraft(r.Context(), userOf(r), r.PathValue("key"), string(b))
		}
		if errors.Is(err, store.ErrInvalidKey) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err != nil {
			httpError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("DELETE /api/draft/{key}", func(w http.ResponseWriter, r *http.Request) {
		if err := a.store.DeleteDraft(r.Context(), userOf(r), r.PathValue("key")); err != nil {
			httpError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /api/run", a.handleRun)
}

func (a *app) codeUnit(kind, id string) (*content.CodeUnit, bool) {
	cs := a.cs()
	switch kind {
	case "exercise":
		if e, ok := cs.Exercises[id]; ok {
			return &e.CodeUnit, true
		}
	case "tutorial":
		if t, ok := cs.Tutorials[id]; ok {
			return &t.CodeUnit, true
		}
	}
	return nil, false
}

type runRequest struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
	Lang string `json:"lang"`
	Mode string `json:"mode"`
	Code string `json:"code"`
	Step int    `json:"step"` // tutorials: run tests for steps 1..Step
}

func (a *app) handleRun(w http.ResponseWriter, r *http.Request) {
	var req runRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cu, ok := a.codeUnit(req.Kind, req.ID)
	if !ok || cu.Support[req.Lang] == nil {
		http.NotFound(w, r)
		return
	}
	job := runner.Job{Lang: req.Lang, Mode: runner.Mode(req.Mode), UserCode: req.Code, Files: cu.Support[req.Lang]}
	if job.Mode != runner.ModeRun {
		job.Mode = runner.ModeTest
	}
	if req.Kind == "tutorial" && req.Step > 0 {
		job.TestFilter = content.StepFilter(req.Lang, req.Step)
	}
	select {
	case a.runSlots <- struct{}{}:
		defer func() { <-a.runSlots }()
	case <-r.Context().Done():
		return
	}
	writeJSON(w, runner.Execute(r.Context(), job))
}

// catalog lists everything the UI can link to, for bookmarks and the library page.
func catalog(cs *content.Store) map[string]any {
	type entry struct {
		ID, Title, Summary, Difficulty string
		Minutes                        int
		Tags                           []string
	}
	list := func(n int) []entry { return make([]entry, 0, n) }
	lessons, exercises, tutorials, quizzes := list(len(cs.Lessons)), list(len(cs.Exercises)), list(len(cs.Tutorials)), list(len(cs.Quizzes))
	for _, l := range cs.Lessons {
		lessons = append(lessons, entry{ID: l.ID, Title: l.Title, Summary: l.Summary})
	}
	for _, e := range cs.Exercises {
		exercises = append(exercises, entry{ID: e.ID, Title: e.Title, Summary: e.Summary, Difficulty: e.Difficulty, Minutes: e.Minutes, Tags: e.Tags})
	}
	for _, t := range cs.Tutorials {
		tutorials = append(tutorials, entry{ID: t.ID, Title: t.Title, Summary: t.Summary, Minutes: t.Minutes})
	}
	for _, q := range cs.Quizzes {
		quizzes = append(quizzes, entry{ID: q.ID, Title: q.Title, Summary: q.Summary})
	}
	for _, l := range [][]entry{lessons, exercises, tutorials, quizzes} {
		sort.Slice(l, func(i, j int) bool { return l[i].Title < l[j].Title })
	}
	return map[string]any{"lessons": lessons, "exercises": exercises, "tutorials": tutorials, "quizzes": quizzes}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(v)
}

func httpError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// runVerify checks that every reference solution compiles and passes all
// tests, and that every starter at least compiles. Exit code 1 on failure.
func runVerify(cs *content.Store, only, onlyLang string) int {
	type unit struct {
		name  string
		cu    *content.CodeUnit
		steps int
	}
	var units []unit
	for id, e := range cs.Exercises {
		units = append(units, unit{"exercise/" + id, &e.CodeUnit, 0})
	}
	for id, t := range cs.Tutorials {
		units = append(units, unit{"tutorial/" + id, &t.CodeUnit, len(t.Steps)})
	}
	sort.Slice(units, func(i, j int) bool { return units[i].name < units[j].name })
	failed := 0
	for _, u := range units {
		if only != "" && !strings.Contains(u.name, only) {
			continue
		}
		for _, lang := range content.Langs {
			if onlyLang != "" && lang != onlyLang {
				continue
			}
			if _, ok := u.cu.Starter[lang]; !ok {
				fmt.Printf("FAIL %-40s %-6s missing (no %s/ folder)\n", u.name, lang, lang)
				failed++
				continue
			}
			ctx := context.Background()
			filter := ""
			if u.steps > 0 {
				filter = content.StepFilter(lang, u.steps)
			}
			ref := runner.Execute(ctx, runner.Job{Lang: lang, Mode: runner.ModeTest, UserCode: u.cu.Reference[lang], Files: u.cu.Support[lang], TestFilter: filter, Timeout: 90 * time.Second})
			run := runner.Execute(ctx, runner.Job{Lang: lang, Mode: runner.ModeRun, UserCode: u.cu.Reference[lang], Files: u.cu.Support[lang], Timeout: 60 * time.Second})
			starter := runner.Execute(ctx, runner.Job{Lang: lang, Mode: runner.ModeTest, UserCode: u.cu.Starter[lang], Files: u.cu.Support[lang], TestFilter: filter, Timeout: 90 * time.Second})
			ok := ref.Status == "ok" && run.Status == "ok" && starter.Status != "compile_error" && starter.Status != "ok"
			mark := "PASS"
			if !ok {
				mark = "FAIL"
				failed++
			}
			fmt.Printf("%s %-40s %-6s ref=%s run=%s starter=%s\n", mark, u.name, lang, ref.Describe(), run.Status, starter.Describe())
			if !ok {
				for _, t := range ref.Tests {
					if !t.Passed {
						fmt.Printf("     ✗ %s: %s\n", t.Name, t.Message)
					}
				}
				for _, r := range []runner.Result{ref, run, starter} {
					if r.Status == "compile_error" || r.Status == "runtime_error" || r.Status == "timeout" {
						fmt.Printf("     %s\n%s\n", r.Status, indent(r.Stderr+r.Stdout))
					}
				}
			}
		}
	}
	if failed > 0 {
		fmt.Printf("\n%d failure(s)\n", failed)
		return 1
	}
	fmt.Println("\nall reference solutions pass")
	return 0
}

func indent(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) > 25 {
		lines = append(lines[:25], "...")
	}
	return "       " + strings.Join(lines, "\n       ")
}
