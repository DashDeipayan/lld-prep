package content

import (
	"os"
	"strings"
	"testing"
)

// The real course content must load, and every practice step must have both
// a task and a model answer.
func TestPracticesLoad(t *testing.T) {
	s, err := Load(os.DirFS("../../content"))
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"parking-lot", "music-player", "risk-management", "notification-system"} {
		p, ok := s.Practices[id]
		if !ok {
			t.Fatalf("practice %s missing", id)
		}
		for i, st := range p.Steps {
			if st.Task == "" || st.Answer == "" {
				t.Errorf("%s step %d: empty task or answer", id, i+1)
			}
			if strings.Contains(st.Answer, answerMarker) {
				t.Errorf("%s step %d: more than one answer marker", id, i+1)
			}
		}
		for _, l := range p.Related {
			if l.Title == "" {
				t.Errorf("%s: related %s/%s not resolved", id, l.Type, l.ID)
			}
		}
	}
	if hits := s.Search("shuffle"); len(hits) == 0 || hits[0].Type != "practice" {
		t.Errorf("search for shuffle: %+v", hits)
	}
}
