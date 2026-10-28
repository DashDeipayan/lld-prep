package main

import "errors"

var ErrOutOfRange = errors.New("position out of range")

// Suggested shape (you may change internals, not the TextEditor API):
type Command interface {
	Execute()
	Undo()
}

type TextEditor struct {
	// TODO
}

func NewTextEditor(maxHistory int) *TextEditor { panic("TODO") }

func (e *TextEditor) Text() string                   { panic("TODO") }
func (e *TextEditor) Insert(pos int, s string) error { panic("TODO") }
func (e *TextEditor) Delete(from, to int) error      { panic("TODO") }
func (e *TextEditor) Undo() bool                     { panic("TODO") }
func (e *TextEditor) Redo() bool                     { panic("TODO") }
