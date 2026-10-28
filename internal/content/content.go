// Package content loads the course material (plan, lessons, exercises,
// tutorials, quizzes, mocks) from a file system and builds a search index.
package content

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

type PlanItem struct {
	Type    string `json:"type"` // lesson | exercise | tutorial | quiz | mock
	ID      string `json:"id"`
	Minutes int    `json:"minutes"`
	Title   string `json:"title,omitempty"` // filled in by the loader
	Note    string `json:"note,omitempty"`
}

type Day struct {
	Day   int        `json:"day"`
	Title string     `json:"title"`
	Goal  string     `json:"goal"`
	Items []PlanItem `json:"items"`
}

type Plan struct {
	Days []Day `json:"days"`
}

type Lesson struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Markdown string `json:"markdown"`
}

// CodeUnit is the shared shape of exercises and tutorials: per-language
// starter code, reference solution and support files (driver + tests).
type CodeUnit struct {
	Starter   map[string]string            `json:"starter"`
	Reference map[string]string            `json:"-"`
	Support   map[string]map[string][]byte `json:"-"`
}

type Exercise struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Difficulty string   `json:"difficulty"`
	Summary    string   `json:"summary"`
	Minutes    int      `json:"minutes"`
	Tags       []string `json:"tags"`
	Hints      []string `json:"hints"`
	Problem    string   `json:"problem"`
	CodeUnit
}

type TutorialStep struct {
	Title    string `json:"title"`
	Doc      string `json:"doc,omitempty"`
	Markdown string `json:"markdown"`
}

type Tutorial struct {
	ID      string         `json:"id"`
	Title   string         `json:"title"`
	Summary string         `json:"summary"`
	Minutes int            `json:"minutes"`
	Intro   string         `json:"intro"`
	Steps   []TutorialStep `json:"steps"`
	CodeUnit
}

type Question struct {
	Q       string   `json:"q"`
	Code    string   `json:"code,omitempty"`
	Options []string `json:"options"`
	Answer  int      `json:"answer"`
	Explain string   `json:"explain"`
}

type Quiz struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Summary   string     `json:"summary"`
	Questions []Question `json:"questions"`
}

type Mock struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Minutes  int    `json:"minutes"`
	Exercise string `json:"exercise"`
	Brief    string `json:"brief"`
	Company  string `json:"company"`
}

type RubricItem struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Weight   int    `json:"weight"`
	Strong   string `json:"strong"`
	Weak     string `json:"weak"`
}

type Store struct {
	Plan      Plan
	Lessons   map[string]*Lesson
	Exercises map[string]*Exercise
	Tutorials map[string]*Tutorial
	Quizzes   map[string]*Quiz
	Mocks     []*Mock
	Rubric    []RubricItem
	index     []doc
}

var langs = []string{"java", "go"}

// Load reads everything from fsys (rooted at the content directory).
func Load(fsys fs.FS) (*Store, error) {
	s := &Store{
		Lessons:   map[string]*Lesson{},
		Exercises: map[string]*Exercise{},
		Tutorials: map[string]*Tutorial{},
		Quizzes:   map[string]*Quiz{},
	}
	if err := readJSON(fsys, "plan.json", &s.Plan); err != nil {
		return nil, err
	}
	if err := readJSON(fsys, "rubric.json", &s.Rubric); err != nil {
		return nil, err
	}
	if err := s.loadLessons(fsys); err != nil {
		return nil, err
	}
	if err := s.loadExercises(fsys); err != nil {
		return nil, err
	}
	if err := s.loadTutorials(fsys); err != nil {
		return nil, err
	}
	if err := s.loadQuizzes(fsys); err != nil {
		return nil, err
	}
	if err := s.loadMocks(fsys); err != nil {
		return nil, err
	}
	if err := s.resolvePlan(); err != nil {
		return nil, err
	}
	s.buildIndex()
	return s, nil
}

func readJSON(fsys fs.FS, name string, v any) error {
	b, err := fs.ReadFile(fsys, name)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

// parseFrontMatter splits "---\nkey: value\n---\nbody".
func parseFrontMatter(src string) (map[string]string, string) {
	meta := map[string]string{}
	src = strings.ReplaceAll(src, "\r\n", "\n")
	if !strings.HasPrefix(src, "---\n") {
		return meta, src
	}
	end := strings.Index(src[4:], "\n---\n")
	if end < 0 {
		return meta, src
	}
	for _, line := range strings.Split(src[4:4+end], "\n") {
		if k, v, ok := strings.Cut(line, ":"); ok {
			meta[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return meta, src[4+end+5:]
}

func (s *Store) loadLessons(fsys fs.FS) error {
	entries, err := fs.ReadDir(fsys, "lessons")
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		b, err := fs.ReadFile(fsys, path.Join("lessons", e.Name()))
		if err != nil {
			return err
		}
		meta, body := parseFrontMatter(string(b))
		id := strings.TrimSuffix(e.Name(), ".md")
		s.Lessons[id] = &Lesson{ID: id, Title: meta["title"], Summary: meta["summary"], Markdown: body}
	}
	return nil
}

func loadCodeUnit(fsys fs.FS, dir string) (CodeUnit, error) {
	cu := CodeUnit{Starter: map[string]string{}, Reference: map[string]string{}, Support: map[string]map[string][]byte{}}
	for _, lang := range langs {
		ldir := path.Join(dir, lang)
		entries, err := fs.ReadDir(fsys, ldir)
		if err != nil {
			return cu, fmt.Errorf("%s: %w", ldir, err)
		}
		support := map[string][]byte{}
		for _, e := range entries {
			b, err := fs.ReadFile(fsys, path.Join(ldir, e.Name()))
			if err != nil {
				return cu, err
			}
			switch e.Name() {
			case "Solution.java", "solution.go":
				cu.Starter[lang] = string(b)
			case "Reference.java", "reference.go.txt":
				cu.Reference[lang] = string(b)
			default:
				support[e.Name()] = b
			}
		}
		if cu.Starter[lang] == "" || cu.Reference[lang] == "" {
			return cu, fmt.Errorf("%s: missing starter or reference", ldir)
		}
		cu.Support[lang] = support
	}
	return cu, nil
}

func (s *Store) loadExercises(fsys fs.FS) error {
	entries, err := fs.ReadDir(fsys, "exercises")
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := path.Join("exercises", e.Name())
		ex := &Exercise{}
		if err := readJSON(fsys, path.Join(dir, "exercise.json"), ex); err != nil {
			return err
		}
		ex.ID = e.Name()
		p, err := fs.ReadFile(fsys, path.Join(dir, "problem.md"))
		if err != nil {
			return err
		}
		ex.Problem = string(p)
		if ex.CodeUnit, err = loadCodeUnit(fsys, dir); err != nil {
			return err
		}
		s.Exercises[ex.ID] = ex
	}
	return nil
}

func (s *Store) loadTutorials(fsys fs.FS) error {
	entries, err := fs.ReadDir(fsys, "tutorials")
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := path.Join("tutorials", e.Name())
		t := &Tutorial{}
		if err := readJSON(fsys, path.Join(dir, "tutorial.json"), t); err != nil {
			return err
		}
		t.ID = e.Name()
		intro, err := fs.ReadFile(fsys, path.Join(dir, "intro.md"))
		if err != nil {
			return err
		}
		t.Intro = string(intro)
		for i := range t.Steps {
			b, err := fs.ReadFile(fsys, path.Join(dir, t.Steps[i].Doc))
			if err != nil {
				return err
			}
			t.Steps[i].Markdown = string(b)
		}
		if t.CodeUnit, err = loadCodeUnit(fsys, dir); err != nil {
			return err
		}
		s.Tutorials[t.ID] = t
	}
	return nil
}

func (s *Store) loadQuizzes(fsys fs.FS) error {
	entries, err := fs.ReadDir(fsys, "quizzes")
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		q := &Quiz{}
		if err := readJSON(fsys, path.Join("quizzes", e.Name()), q); err != nil {
			return err
		}
		q.ID = strings.TrimSuffix(e.Name(), ".json")
		for i, qq := range q.Questions {
			if qq.Answer < 0 || qq.Answer >= len(qq.Options) {
				return fmt.Errorf("quiz %s question %d: answer out of range", q.ID, i+1)
			}
		}
		s.Quizzes[q.ID] = q
	}
	return nil
}

func (s *Store) loadMocks(fsys fs.FS) error {
	if err := readJSON(fsys, "mocks/mocks.json", &s.Mocks); err != nil {
		return err
	}
	for _, m := range s.Mocks {
		if _, ok := s.Exercises[m.Exercise]; !ok {
			return fmt.Errorf("mock %s: unknown exercise %s", m.ID, m.Exercise)
		}
	}
	return nil
}

// Title returns the display title of any plan item.
func (s *Store) Title(typ, id string) (string, bool) {
	switch typ {
	case "lesson":
		if l, ok := s.Lessons[id]; ok {
			return l.Title, true
		}
	case "exercise":
		if e, ok := s.Exercises[id]; ok {
			return e.Title, true
		}
	case "tutorial":
		if t, ok := s.Tutorials[id]; ok {
			return t.Title, true
		}
	case "quiz":
		if q, ok := s.Quizzes[id]; ok {
			return q.Title, true
		}
	case "mock":
		for _, m := range s.Mocks {
			if m.ID == id {
				return m.Title, true
			}
		}
	}
	return "", false
}

func (s *Store) resolvePlan() error {
	for di := range s.Plan.Days {
		for ii := range s.Plan.Days[di].Items {
			it := &s.Plan.Days[di].Items[ii]
			t, ok := s.Title(it.Type, it.ID)
			if !ok {
				return fmt.Errorf("plan day %d: unknown %s %q", s.Plan.Days[di].Day, it.Type, it.ID)
			}
			it.Title = t
		}
	}
	return nil
}

// StepFilter returns the test-name regex that covers steps 1..n.
func StepFilter(lang string, n int) string {
	parts := make([]string, n)
	for i := range parts {
		if lang == "go" {
			parts[i] = fmt.Sprintf("S%d", i+1)
		} else {
			parts[i] = fmt.Sprintf("s%d", i+1)
		}
	}
	if lang == "go" {
		return "^Test(" + strings.Join(parts, "|") + ")_"
	}
	return "^(" + strings.Join(parts, "|") + ")_"
}

// ---- search ----

type doc struct {
	Type, ID, Title, Text string
	lowerTitle, lower     string
}

type Hit struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	score   int
}

func (s *Store) buildIndex() {
	add := func(typ, id, title, text string) {
		s.index = append(s.index, doc{Type: typ, ID: id, Title: title, Text: text,
			lowerTitle: strings.ToLower(title), lower: strings.ToLower(title + "\n" + text)})
	}
	for _, l := range s.Lessons {
		add("lesson", l.ID, l.Title, l.Summary+"\n"+l.Markdown)
	}
	for _, e := range s.Exercises {
		add("exercise", e.ID, e.Title, e.Summary+"\n"+strings.Join(e.Tags, " ")+"\n"+e.Problem)
	}
	for _, t := range s.Tutorials {
		var b strings.Builder
		b.WriteString(t.Summary + "\n" + t.Intro)
		for _, st := range t.Steps {
			b.WriteString("\n" + st.Title + "\n" + st.Markdown)
		}
		add("tutorial", t.ID, t.Title, b.String())
	}
	for _, q := range s.Quizzes {
		var b strings.Builder
		for _, qq := range q.Questions {
			b.WriteString(qq.Q + "\n" + qq.Explain + "\n")
		}
		add("quiz", q.ID, q.Title, b.String())
	}
}

// Search returns documents containing every query term, best first.
func (s *Store) Search(q string) []Hit {
	terms := strings.Fields(strings.ToLower(q))
	if len(terms) == 0 {
		return nil
	}
	var hits []Hit
	for _, d := range s.index {
		score := 0
		ok := true
		for _, t := range terms {
			c := strings.Count(d.lower, t)
			if c == 0 {
				ok = false
				break
			}
			score += min(c, 20)
			if strings.Contains(d.lowerTitle, t) {
				score += 50
			}
		}
		if !ok {
			continue
		}
		hits = append(hits, Hit{Type: d.Type, ID: d.ID, Title: d.Title, Snippet: snippet(d.Text, terms[0]), score: score})
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].score > hits[j].score })
	if len(hits) > 30 {
		hits = hits[:30]
	}
	return hits
}

func snippet(text, term string) string {
	clean := strings.Join(strings.Fields(stripMarkdown(text)), " ")
	i := strings.Index(strings.ToLower(clean), term)
	if i < 0 {
		i = 0
	}
	start := max(0, i-80)
	end := min(len(clean), i+len(term)+120)
	out := clean[start:end]
	if start > 0 {
		out = "…" + out
	}
	if end < len(clean) {
		out += "…"
	}
	return strings.ToValidUTF8(out, "")
}

func stripMarkdown(s string) string {
	r := strings.NewReplacer("#", "", "*", "", "`", "", "|", " ", ">", "", "_", " ")
	return r.Replace(s)
}
