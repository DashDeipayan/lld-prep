package main

import "errors"

var (
	ErrTopicExists   = errors.New("topic exists")
	ErrUnknownTopic  = errors.New("unknown topic")
	ErrUnknownGroup  = errors.New("unknown group")
	ErrInvalidOffset = errors.New("invalid offset")
)

type Message struct {
	Offset  int64
	Payload string
}

type Broker struct {
	// TODO
}

func NewBroker() *Broker { panic("TODO") }

func (b *Broker) CreateTopic(t string) error                          { panic("TODO") }
func (b *Broker) Publish(t, payload string) (int64, error)            { panic("TODO") }
func (b *Broker) Subscribe(t, group string, fromBeginning bool) error { panic("TODO") }
func (b *Broker) Poll(t, group string, max int) ([]Message, error)    { panic("TODO") }
func (b *Broker) Seek(t, group string, offset int64) error            { panic("TODO") }
func (b *Broker) Lag(t, group string) (int64, error)                  { panic("TODO") }
