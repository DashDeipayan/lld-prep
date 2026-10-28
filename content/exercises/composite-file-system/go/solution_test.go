package main

import (
	"errors"
	"reflect"
	"testing"
)

func sample(t *testing.T) *FileSystem {
	t.Helper()
	fs := NewFileSystem()
	for _, err := range []error{
		fs.Mkdir("/home/alice/docs"),
		fs.CreateFile("/home/alice/docs/cv.pdf", 120),
		fs.CreateFile("/home/alice/docs/notes.txt", 30),
		fs.Mkdir("/home/bob"),
		fs.CreateFile("/home/bob/todo.txt", 5),
		fs.CreateFile("/readme.txt", 1),
	} {
		if err != nil {
			t.Fatalf("setup: %v", err)
		}
	}
	return fs
}

func strs(t *testing.T, want []string, got []string, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: unexpected error %v", msg, err)
	}
	if len(want) == 0 && len(got) == 0 {
		return
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s\nexpected: %v\n  actual: %v", msg, want, got)
	}
}

func size(t *testing.T, fs *FileSystem, p string, want int64) {
	t.Helper()
	got, err := fs.Size(p)
	if err != nil || got != want {
		t.Errorf("size(%s): expected %d, got %d (err %v)", p, want, got, err)
	}
}

func errIs(t *testing.T, err, target error, msg string) {
	t.Helper()
	if !errors.Is(err, target) {
		t.Errorf("%s: expected %v, got %v", msg, target, err)
	}
}

func TestMkdirPAndLs(t *testing.T) {
	defer guard(t)
	fs := sample(t)
	got, err := fs.Ls("/")
	strs(t, []string{"home/", "readme.txt"}, got, err, "root listing")
	got, err = fs.Ls("/home")
	strs(t, []string{"alice/", "bob/"}, got, err, "nested")
	got, err = fs.Ls("/home/alice/docs/cv.pdf")
	strs(t, []string{"cv.pdf"}, got, err, "ls on a file")
	if err := fs.Mkdir("/home/alice"); err != nil {
		t.Errorf("mkdir on existing dir should succeed: %v", err)
	}
}

func TestRecursiveSize(t *testing.T) {
	defer guard(t)
	fs := sample(t)
	size(t, fs, "/", 156)
	size(t, fs, "/home", 155)
	size(t, fs, "/home/alice/docs/notes.txt", 30)
	fs.Mkdir("/empty")
	size(t, fs, "/empty", 0)
}

func TestFindByExtension(t *testing.T) {
	defer guard(t)
	fs := sample(t)
	got, err := fs.Find("/", "txt")
	strs(t, []string{"/home/alice/docs/notes.txt", "/home/bob/todo.txt", "/readme.txt"}, got, err, "all txt, sorted")
	got, err = fs.Find("/home/bob", "txt")
	strs(t, []string{"/home/bob/todo.txt"}, got, err, "scoped")
	got, err = fs.Find("/home", "png")
	strs(t, nil, got, err, "none")
}

func TestDelete(t *testing.T) {
	defer guard(t)
	fs := sample(t)
	if err := fs.Delete("/home/alice"); err != nil {
		t.Fatal(err)
	}
	got, err := fs.Ls("/home")
	strs(t, []string{"bob/"}, got, err, "subtree removed")
	size(t, fs, "/", 6)
	fs.Delete("/readme.txt")
	_, err = fs.Size("/readme.txt")
	errIs(t, err, ErrNotFound, "file gone")
	errIs(t, fs.Delete("/"), ErrInvalidPath, "cannot delete root")
	errIs(t, fs.Delete("/nope"), ErrNotFound, "delete missing")
}

func TestErrors(t *testing.T) {
	defer guard(t)
	fs := sample(t)
	errIs(t, fs.CreateFile("/missing/x.txt", 1), ErrNotFound, "parent missing")
	errIs(t, fs.CreateFile("/readme.txt/x", 1), ErrNotDir, "parent is a file")
	errIs(t, fs.CreateFile("/home/bob/todo.txt", 9), ErrExists, "file exists")
	errIs(t, fs.Mkdir("/readme.txt/sub"), ErrExists, "mkdir through a file")
	_, err := fs.Ls("/nope")
	errIs(t, err, ErrNotFound, "ls missing")
	_, err = fs.Find("/nope", "txt")
	errIs(t, err, ErrNotFound, "find missing")
}
