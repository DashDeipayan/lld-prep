package main

import "errors"

// ---- Provided ----

type User struct {
	ID          string
	Email       string
	Phone       string
	DeviceToken string
	Muted       map[string]bool
}

type Outbox struct{ Sent []string }

func (o *Outbox) Record(channel, to, msg string) { o.Sent = append(o.Sent, channel+":"+to+":"+msg) }

var ErrEmptyMessage = errors.New("message must not be blank")

// ---- The code you are replacing (read it, then delete it) ----

// LegacyNotify: new channel => edit this function; one failure stops the rest.
func LegacyNotify(u User, msg string, o *Outbox) int {
	sent := 0
	if u.Email != "" && !u.Muted["email"] {
		o.Record("email", u.Email, msg)
		sent++
	}
	if u.Phone != "" && !u.Muted["sms"] {
		o.Record("sms", u.Phone, msg)
		sent++
	}
	if u.DeviceToken != "" && !u.Muted["push"] {
		o.Record("push", u.DeviceToken, msg)
		sent++
	}
	return sent
}

// ---- Your design ----

type Notifier interface {
	Channel() string
	CanReach(u User) bool
	Send(u User, msg string) error
}

func NewEmailNotifier(o *Outbox) Notifier { panic("TODO") }
func NewSmsNotifier(o *Outbox) Notifier   { panic("TODO") }
func NewPushNotifier(o *Outbox) Notifier  { panic("TODO") }

type NotificationService struct {
	// TODO
}

func NewNotificationService(ns ...Notifier) *NotificationService { panic("TODO") }

// Notify returns the number of successful sends.
func (s *NotificationService) Notify(u User, msg string) (int, error) { panic("TODO") }
