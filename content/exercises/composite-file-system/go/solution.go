package main

import "errors"

var (
	ErrNotFound    = errors.New("not found")
	ErrExists      = errors.New("already exists")
	ErrNotDir      = errors.New("not a directory")
	ErrInvalidPath = errors.New("invalid path")
)

// TODO: a Node interface with File and Dir implementations (Composite)

type FileSystem struct {
	// TODO
}

func NewFileSystem() *FileSystem { panic("TODO") }

func (fs *FileSystem) Mkdir(p string) error                  { panic("TODO") }
func (fs *FileSystem) CreateFile(p string, size int64) error { panic("TODO") }
func (fs *FileSystem) Size(p string) (int64, error)          { panic("TODO") }
func (fs *FileSystem) Ls(p string) ([]string, error)         { panic("TODO") }
func (fs *FileSystem) Find(p, ext string) ([]string, error)  { panic("TODO") }
func (fs *FileSystem) Delete(p string) error                 { panic("TODO") }
