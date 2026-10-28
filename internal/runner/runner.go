// Package runner compiles and executes user code with the locally installed
// JDK and Go toolchains, and collects structured test results.
package runner

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

//go:embed assets/TestRunner.java
var javaTestRunner []byte

//go:embed assets/guard_test.go.txt
var goGuard []byte

// Mode selects what is executed after compilation.
type Mode string

const (
	ModeRun  Mode = "run"  // run the driver (Main.java / main.go) and show stdout
	ModeTest Mode = "test" // run the hidden test-suite
)

// Job describes one execution request.
type Job struct {
	Lang string // "java" or "go"
	Mode Mode
	// UserCode replaces the starter file (Solution.java / solution.go).
	UserCode string
	// Files are the support files for the language (driver, tests, ...),
	// keyed by file name. The user file is added by the runner.
	Files map[string][]byte
	// TestFilter is a regex of test names to run; empty means all.
	TestFilter string
	Timeout    time.Duration
}

// TestResult is the outcome of a single test case.
type TestResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message,omitempty"`
}

// Result is returned to the UI.
type Result struct {
	Status     string       `json:"status"` // ok | compile_error | runtime_error | timeout | failed | toolchain_missing
	Stdout     string       `json:"stdout"`
	Stderr     string       `json:"stderr"`
	Tests      []TestResult `json:"tests,omitempty"`
	Passed     int          `json:"passed"`
	Total      int          `json:"total"`
	DurationMs int64        `json:"durationMs"`
}

// UserFileName returns the name of the editable file for a language.
func UserFileName(lang string) string {
	if lang == "go" {
		return "solution.go"
	}
	return "Solution.java"
}

const maxOutput = 64 * 1024

// Execute runs the job in a fresh temporary directory.
func Execute(ctx context.Context, job Job) Result {
	start := time.Now()
	var res Result
	switch job.Lang {
	case "java":
		res = runJava(ctx, job)
	case "go":
		res = runGo(ctx, job)
	default:
		res = Result{Status: "runtime_error", Stderr: "unsupported language " + job.Lang}
	}
	res.DurationMs = time.Since(start).Milliseconds()
	res.Stdout = truncate(res.Stdout)
	res.Stderr = truncate(res.Stderr)
	return res
}

func truncate(s string) string {
	if len(s) > maxOutput {
		return s[:maxOutput] + "\n... output truncated ..."
	}
	return s
}

func workspace(job Job) (string, error) {
	dir, err := os.MkdirTemp("", "lldprep-*")
	if err != nil {
		return "", err
	}
	for name, data := range job.Files {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			return dir, err
		}
	}
	err = os.WriteFile(filepath.Join(dir, UserFileName(job.Lang)), []byte(job.UserCode), 0o644)
	return dir, err
}

func timeoutOf(job Job) time.Duration {
	if job.Timeout > 0 {
		return job.Timeout
	}
	return 20 * time.Second
}

// command runs a process and reports whether the deadline was hit.
func command(ctx context.Context, dir string, env []string, name string, args ...string) (stdout, stderr string, exitErr error, timedOut bool) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}
	setProcessGroup(cmd)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return out.String(), errb.String(), err, true
	}
	return out.String(), errb.String(), err, false
}

func runJava(parent context.Context, job Job) Result {
	if _, err := exec.LookPath("javac"); err != nil {
		return Result{Status: "toolchain_missing", Stderr: "javac not found on PATH. Install JDK 17+ to run Java code."}
	}
	if job.Mode == ModeTest {
		job.Files = withFile(job.Files, "TestRunner.java", javaTestRunner)
	} else {
		// Tests.java depends on TestRunner.java helpers; leave it out of plain runs.
		files := map[string][]byte{}
		for k, v := range job.Files {
			if k != "Tests.java" {
				files[k] = v
			}
		}
		job.Files = files
	}
	dir, err := workspace(job)
	defer os.RemoveAll(dir)
	if err != nil {
		return Result{Status: "runtime_error", Stderr: err.Error()}
	}
	ctx, cancel := context.WithTimeout(parent, timeoutOf(job))
	defer cancel()

	var sources []string
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".java") {
			sources = append(sources, e.Name())
		}
	}
	args := append([]string{"-d", "out", "-encoding", "UTF-8", "-Xlint:none"}, sources...)
	_, cerr, err, to := command(ctx, dir, nil, "javac", args...)
	if to {
		return Result{Status: "timeout", Stderr: "compilation timed out"}
	}
	if err != nil {
		return Result{Status: "compile_error", Stderr: cerr}
	}

	if job.Mode == ModeRun {
		out, serr, err, to := command(ctx, dir, nil, "java", "-cp", "out", "Main")
		return finishRun(out, serr, err, to)
	}
	filter := job.TestFilter
	if filter == "" {
		filter = "^test"
	}
	out, serr, err, to := command(ctx, dir, nil, "java", "-cp", "out", "TestRunner", filter)
	tests, plain := parseMarkers(out)
	return finishTests(tests, plain, serr, err, to)
}

func runGo(parent context.Context, job Job) Result {
	if _, err := exec.LookPath("go"); err != nil {
		return Result{Status: "toolchain_missing", Stderr: "go not found on PATH. Install Go 1.23+ to run Go code."}
	}
	job.Files = withFile(job.Files, "go.mod", []byte("module solution\n\ngo 1.21\n"))
	if job.Mode == ModeRun {
		// Tests are not needed (and would slow vet) for a plain run.
		files := map[string][]byte{}
		for k, v := range job.Files {
			if !strings.HasSuffix(k, "_test.go") {
				files[k] = v
			}
		}
		job.Files = files
	} else {
		job.Files = withFile(job.Files, "zz_guard_test.go", goGuard)
	}
	dir, err := workspace(job)
	defer os.RemoveAll(dir)
	if err != nil {
		return Result{Status: "runtime_error", Stderr: err.Error()}
	}
	ctx, cancel := context.WithTimeout(parent, timeoutOf(job))
	defer cancel()
	env := []string{"GOFLAGS=-mod=mod", "GOTOOLCHAIN=local", "GOPROXY=off", "GOWORK=off"}

	if job.Mode == ModeRun {
		out, serr, err, to := command(ctx, dir, env, "go", "run", ".")
		if err != nil && !to && isGoBuildFailure(serr) {
			return Result{Status: "compile_error", Stderr: cleanGoPaths(serr, dir)}
		}
		return finishRun(out, cleanGoPaths(serr, dir), err, to)
	}

	filter := job.TestFilter
	if filter == "" {
		filter = "^Test"
	}
	out, serr, err, to := command(ctx, dir, env, "go", "test", "-json", "-count=1", "-vet=off", "-timeout", "15s", "-run", filter, ".")
	tests, plain, buildFailed := parseGoTestJSON(out)
	if buildFailed || (len(tests) == 0 && err != nil && !to) {
		msg := cleanGoPaths(serr+plain, dir)
		return Result{Status: "compile_error", Stderr: msg}
	}
	return finishTests(tests, plain, cleanGoPaths(serr, dir), err, to)
}

func isGoBuildFailure(stderr string) bool {
	return strings.Contains(stderr, ".go:") && !strings.Contains(stderr, "panic:") && !strings.Contains(stderr, "goroutine ")
}

func cleanGoPaths(s, dir string) string {
	s = strings.ReplaceAll(s, dir+string(os.PathSeparator), "")
	return strings.ReplaceAll(s, dir, ".")
}

func withFile(files map[string][]byte, name string, data []byte) map[string][]byte {
	out := make(map[string][]byte, len(files)+1)
	for k, v := range files {
		out[k] = v
	}
	out[name] = data
	return out
}

func finishRun(out, serr string, err error, timedOut bool) Result {
	r := Result{Status: "ok", Stdout: out, Stderr: serr}
	switch {
	case timedOut:
		r.Status = "timeout"
		r.Stderr += "\nExecution timed out (possible infinite loop or deadlock)."
	case err != nil:
		r.Status = "runtime_error"
	}
	return r
}

func finishTests(tests []TestResult, plain, serr string, err error, timedOut bool) Result {
	r := Result{Tests: tests, Stdout: plain, Stderr: serr, Total: len(tests)}
	for _, t := range tests {
		if t.Passed {
			r.Passed++
		}
	}
	switch {
	case timedOut:
		r.Status = "timeout"
		r.Stderr += "\nTests timed out (possible infinite loop or deadlock)."
	case r.Total == 0:
		r.Status = "runtime_error"
		if strings.TrimSpace(r.Stderr) == "" {
			if err != nil {
				r.Stderr = err.Error()
			} else {
				r.Stderr = "no tests were run"
			}
		}
	case r.Passed == r.Total:
		r.Status = "ok"
	default:
		r.Status = "failed"
	}
	return r
}

// parseMarkers extracts "@@TEST|name|PASS|msg" lines emitted by TestRunner.java.
func parseMarkers(out string) ([]TestResult, string) {
	var tests []TestResult
	var plain strings.Builder
	sc := bufio.NewScanner(strings.NewReader(out))
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "@@TEST|") {
			parts := strings.SplitN(line, "|", 4)
			if len(parts) == 4 {
				msg := strings.ReplaceAll(parts[3], `\n`, "\n")
				tests = append(tests, TestResult{Name: parts[1], Passed: parts[2] == "PASS", Message: msg})
				continue
			}
		}
		plain.WriteString(line)
		plain.WriteByte('\n')
	}
	return tests, plain.String()
}

type goEvent struct {
	Action string
	Test   string
	Output string
}

func parseGoTestJSON(out string) ([]TestResult, string, bool) {
	var tests []TestResult
	outputs := map[string]*strings.Builder{}
	var plain strings.Builder
	buildFailed := false
	sc := bufio.NewScanner(strings.NewReader(out))
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		var ev goEvent
		if err := json.Unmarshal(sc.Bytes(), &ev); err != nil {
			plain.WriteString(sc.Text() + "\n")
			continue
		}
		if ev.Action == "build-output" {
			plain.WriteString(ev.Output)
			buildFailed = true
			continue
		}
		if ev.Action == "build-fail" {
			buildFailed = true
			continue
		}
		if ev.Test == "" {
			if ev.Action == "output" && !isGoNoise(ev.Output) {
				plain.WriteString(ev.Output)
			}
			continue
		}
		top := strings.SplitN(ev.Test, "/", 2)[0]
		b := outputs[top]
		if b == nil {
			b = &strings.Builder{}
			outputs[top] = b
		}
		switch ev.Action {
		case "output":
			if !isGoNoise(ev.Output) {
				b.WriteString(ev.Output)
			}
		case "pass", "fail":
			if strings.Contains(ev.Test, "/") {
				continue // subtests roll up into their parent
			}
			tests = append(tests, TestResult{Name: ev.Test, Passed: ev.Action == "pass", Message: strings.TrimSpace(b.String())})
		}
	}
	return tests, plain.String(), buildFailed
}

func isGoNoise(line string) bool {
	t := strings.TrimSpace(line)
	return strings.HasPrefix(t, "=== RUN") || strings.HasPrefix(t, "--- PASS") || strings.HasPrefix(t, "--- FAIL") ||
		strings.HasPrefix(t, "=== PAUSE") || strings.HasPrefix(t, "=== CONT") ||
		t == "PASS" || t == "FAIL" || strings.HasPrefix(t, "ok  \t") || strings.HasPrefix(t, "FAIL\t") || strings.HasPrefix(t, "ok \t")
}

// Toolchains reports which toolchains are available and their versions.
func Toolchains() map[string]string {
	out := map[string]string{}
	if o, err := exec.Command("javac", "-version").CombinedOutput(); err == nil {
		out["java"] = strings.TrimSpace(string(o))
	}
	if o, err := exec.Command("go", "version").CombinedOutput(); err == nil {
		out["go"] = strings.TrimSpace(string(o))
	}
	return out
}

// Describe is a helper for CLI output.
func (r Result) Describe() string {
	return fmt.Sprintf("%s %d/%d (%dms)", r.Status, r.Passed, r.Total, r.DurationMs)
}
