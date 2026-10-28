package main

import (
	"errors"
	"time"
)

var (
	ErrInvalidURL     = errors.New("invalid url")
	ErrInvalidMethod  = errors.New("invalid method")
	ErrInvalidTimeout = errors.New("invalid timeout")
	ErrBodyNotAllowed = errors.New("body not allowed for this method")
)

type Request struct {
	// TODO: unexported fields
}

// Option configures a Request and may reject bad values.
type Option func(*Request) error

func WithMethod(m string) Option           { panic("TODO") }
func WithHeader(name, value string) Option { panic("TODO") }
func WithTimeout(d time.Duration) Option   { panic("TODO") }
func WithBody(b string) Option             { panic("TODO") }

func NewRequest(url string, opts ...Option) (*Request, error) { panic("TODO") }

func (r *Request) Method() string                    { panic("TODO") }
func (r *Request) URL() string                       { panic("TODO") }
func (r *Request) Header(name string) (string, bool) { panic("TODO") }
func (r *Request) Headers() map[string]string        { panic("TODO") }
func (r *Request) Timeout() time.Duration            { panic("TODO") }
func (r *Request) Body() string                      { panic("TODO") }
