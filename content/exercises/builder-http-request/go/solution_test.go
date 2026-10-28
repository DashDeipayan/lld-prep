package main

import (
	"errors"
	"testing"
	"time"
)

func mustReq(t *testing.T, url string, opts ...Option) *Request {
	t.Helper()
	r, err := NewRequest(url, opts...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return r
}

func wantErr(t *testing.T, target error, url string, opts ...Option) {
	t.Helper()
	_, err := NewRequest(url, opts...)
	if !errors.Is(err, target) {
		t.Errorf("expected %v, got %v", target, err)
	}
}

func TestDefaults(t *testing.T) {
	defer guard(t)
	r := mustReq(t, "https://x.io")
	if r.Method() != "GET" || r.Timeout() != 30*time.Second || len(r.Headers()) != 0 || r.Body() != "" {
		t.Errorf("defaults wrong: method=%q timeout=%v headers=%v body=%q", r.Method(), r.Timeout(), r.Headers(), r.Body())
	}
}

func TestMethodNormalisedAndValidated(t *testing.T) {
	defer guard(t)
	if m := mustReq(t, "http://x.io", WithMethod("patch")).Method(); m != "PATCH" {
		t.Errorf("expected PATCH, got %q", m)
	}
	wantErr(t, ErrInvalidMethod, "http://x.io", WithMethod("FETCH"))
}

func TestHeadersCaseInsensitiveLastWins(t *testing.T) {
	defer guard(t)
	r := mustReq(t, "http://x.io", WithHeader("Content-Type", "text/plain"), WithHeader("content-type", "application/json"))
	if len(r.Headers()) != 1 {
		t.Errorf("expected 1 header, got %v", r.Headers())
	}
	if v, ok := r.Header("CONTENT-TYPE"); !ok || v != "application/json" {
		t.Errorf("lookup should be case-insensitive and last value wins, got %q %v", v, ok)
	}
}

func TestTimeoutBounds(t *testing.T) {
	defer guard(t)
	mustReq(t, "http://x.io", WithTimeout(time.Millisecond))
	mustReq(t, "http://x.io", WithTimeout(time.Minute))
	wantErr(t, ErrInvalidTimeout, "http://x.io", WithTimeout(0))
	wantErr(t, ErrInvalidTimeout, "http://x.io", WithTimeout(time.Minute+time.Millisecond))
}

func TestUrlValidated(t *testing.T) {
	defer guard(t)
	wantErr(t, ErrInvalidURL, "ftp://x.io")
	wantErr(t, ErrInvalidURL, "")
}

func TestBodyRules(t *testing.T) {
	defer guard(t)
	wantErr(t, ErrBodyNotAllowed, "http://x.io", WithBody("x"))
	wantErr(t, ErrBodyNotAllowed, "http://x.io", WithMethod("DELETE"), WithBody("x"))
	if b := mustReq(t, "http://x.io", WithMethod("PUT"), WithBody("x")).Body(); b != "x" {
		t.Errorf("PUT body lost: %q", b)
	}
}

func TestImmutability(t *testing.T) {
	defer guard(t)
	r := mustReq(t, "http://x.io", WithHeader("a", "1"))
	h := r.Headers()
	h["b"] = "2"
	if _, ok := r.Header("b"); ok || len(r.Headers()) != 1 {
		t.Error("mutating the map returned by Headers() must not change the request")
	}
}
