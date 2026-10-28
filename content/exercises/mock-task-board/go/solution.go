package main

import "errors"

type Status int

const (
	Todo Status = iota
	InProgress
	InReview
	Done
)

func (s Status) String() string {
	return [...]string{"TODO", "IN_PROGRESS", "IN_REVIEW", "DONE"}[s]
}

var (
	ErrInvalid           = errors.New("invalid request")
	ErrNotFound          = errors.New("not found")
	ErrIllegalTransition = errors.New("illegal transition")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrBlocked           = errors.New("blocked by dependencies")
	ErrCycle             = errors.New("dependency cycle")
)

type TaskBoard struct {
	// TODO
}

func NewTaskBoard() *TaskBoard { panic("TODO") }

func (b *TaskBoard) Create(title, reporter string, priority int) (string, error) { panic("TODO") }
func (b *TaskBoard) Assign(id, user string) error                                { panic("TODO") }
func (b *TaskBoard) AddDependency(id, dependsOn string) error                    { panic("TODO") }
func (b *TaskBoard) Move(id string, to Status, actor string) error               { panic("TODO") }
func (b *TaskBoard) Status(id string) (Status, error)                            { panic("TODO") }
func (b *TaskBoard) TasksFor(user string) []string                               { panic("TODO") }
func (b *TaskBoard) BlockedTasks() []string                                      { panic("TODO") }
func (b *TaskBoard) History(id string) ([]string, error)                         { panic("TODO") }
