package main

import (
	"errors"
	"reflect"
	"testing"
)

func full(muted ...string) User {
	m := map[string]bool{}
	for _, c := range muted {
		m[c] = true
	}
	return User{ID: "u1", Email: "a@x.com", Phone: "+1555", DeviceToken: "tok", Muted: m}
}

func svc(o *Outbox) *NotificationService {
	return NewNotificationService(NewEmailNotifier(o), NewSmsNotifier(o), NewPushNotifier(o))
}

func sameList(t *testing.T, want, got []string, msg string) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s\nexpected: %v\n  actual: %v", msg, want, got)
	}
}

func count(t *testing.T, want int, got int, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: unexpected error %v", msg, err)
	}
	if want != got {
		t.Errorf("%s\nexpected: %d\n  actual: %d", msg, want, got)
	}
}

func TestChannelsAndReachability(t *testing.T) {
	defer guard(t)
	o := &Outbox{}
	if NewEmailNotifier(o).Channel() != "email" || NewSmsNotifier(o).Channel() != "sms" || NewPushNotifier(o).Channel() != "push" {
		t.Fatal("channel names must be email, sms, push")
	}
	noPhone := User{Email: "a@x.com", DeviceToken: "tok"}
	if !NewEmailNotifier(o).CanReach(noPhone) {
		t.Error("email should be reachable")
	}
	if NewSmsNotifier(o).CanReach(noPhone) {
		t.Error("empty phone -> sms cannot reach")
	}
	if NewPushNotifier(o).CanReach(User{}) {
		t.Error("empty token -> push cannot reach")
	}
}

func TestSendsOnAllReachableChannelsInOrder(t *testing.T) {
	defer guard(t)
	o := &Outbox{}
	n, err := svc(o).Notify(full(), "hi")
	count(t, 3, n, err, "three channels")
	sameList(t, []string{"email:a@x.com:hi", "sms:+1555:hi", "push:tok:hi"}, o.Sent, "outbox contents and order")
}

func TestRespectsMutedChannels(t *testing.T) {
	defer guard(t)
	o := &Outbox{}
	n, err := svc(o).Notify(full("sms"), "hi")
	count(t, 2, n, err, "sms muted")
	sameList(t, []string{"email:a@x.com:hi", "push:tok:hi"}, o.Sent, "no sms recorded")
}

func TestSkipsUnreachable(t *testing.T) {
	defer guard(t)
	o := &Outbox{}
	n, err := svc(o).Notify(User{DeviceToken: "tok"}, "hi")
	count(t, 1, n, err, "only push reachable")
}

type fakeNotifier struct {
	channel string
	fail    bool
	got     *[]string
}

func (f fakeNotifier) Channel() string    { return f.channel }
func (f fakeNotifier) CanReach(User) bool { return true }
func (f fakeNotifier) Send(_ User, m string) error {
	if f.fail {
		return errors.New(f.channel + " down")
	}
	if f.got != nil {
		*f.got = append(*f.got, m)
	}
	return nil
}

func TestFailureIsIsolated(t *testing.T) {
	defer guard(t)
	o := &Outbox{}
	s := NewNotificationService(NewEmailNotifier(o), fakeNotifier{channel: "slack", fail: true}, NewPushNotifier(o))
	n, err := s.Notify(full(), "hi")
	count(t, 2, n, err, "broken channel counted as not sent")
	sameList(t, []string{"email:a@x.com:hi", "push:tok:hi"}, o.Sent, "channels after the broken one still ran")
}

func TestNewChannelWithoutEditingService(t *testing.T) {
	defer guard(t)
	var got []string
	wa := fakeNotifier{channel: "whatsapp", got: &got}
	n, err := NewNotificationService(wa).Notify(full(), "yo")
	count(t, 1, n, err, "custom notifier works")
	n, err = NewNotificationService(wa).Notify(full("whatsapp"), "yo")
	count(t, 0, n, err, "muting applies to custom channels too")
}

func TestRejectsBlankMessage(t *testing.T) {
	defer guard(t)
	_, err := svc(&Outbox{}).Notify(full(), "  ")
	if !errors.Is(err, ErrEmptyMessage) {
		t.Errorf("expected ErrEmptyMessage, got %v", err)
	}
}
