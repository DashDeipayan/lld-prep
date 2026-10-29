from lldtest import check, eq, raises
from solution import EmailNotifier, NotificationService, Notifier, Outbox, PushNotifier, SmsNotifier, User


def full(muted=frozenset()):
    return User("u1", "a@x.com", "+1555", "tok", frozenset(muted))


def svc(o):
    return NotificationService([EmailNotifier(o), SmsNotifier(o), PushNotifier(o)])


def test_channels_and_reachability():
    o = Outbox()
    eq("email", EmailNotifier(o).channel(), "email channel name")
    eq("sms", SmsNotifier(o).channel(), "sms channel name")
    eq("push", PushNotifier(o).channel(), "push channel name")
    no_phone = User("u", "a@x.com", "", "tok")
    check(EmailNotifier(o).can_reach(no_phone), "email reachable")
    check(not SmsNotifier(o).can_reach(no_phone), "empty phone -> sms cannot reach")
    nones = User("u", None, None, None)
    check(not PushNotifier(o).can_reach(nones), "None token -> push cannot reach")


def test_sends_on_all_reachable_channels_in_order():
    o = Outbox()
    eq(3, svc(o).notify(full(), "hi"), "three channels")
    eq(["email:a@x.com:hi", "sms:+1555:hi", "push:tok:hi"], o.sent, "outbox contents and order")


def test_respects_muted_channels():
    o = Outbox()
    eq(2, svc(o).notify(full({"sms"}), "hi"), "sms muted")
    eq(["email:a@x.com:hi", "push:tok:hi"], o.sent, "no sms recorded")


def test_skips_unreachable():
    o = Outbox()
    only_push = User("u", "", None, "tok")
    eq(1, svc(o).notify(only_push, "hi"), "only push reachable")


def test_failure_is_isolated():
    o = Outbox()

    class Broken(Notifier):
        def channel(self):
            return "slack"

        def can_reach(self, user):
            return True

        def send(self, user, message):
            raise RuntimeError("slack down")

    s = NotificationService([EmailNotifier(o), Broken(), PushNotifier(o)])
    eq(2, s.notify(full(), "hi"), "broken channel counted as not sent")
    eq(["email:a@x.com:hi", "push:tok:hi"], o.sent, "channels after the broken one still ran")


def test_new_channel_without_editing_service():
    got = []

    class WhatsApp(Notifier):
        def channel(self):
            return "whatsapp"

        def can_reach(self, user):
            return True

        def send(self, user, message):
            got.append(message)

    eq(1, NotificationService([WhatsApp()]).notify(full(), "yo"), "custom notifier works")
    eq(0, NotificationService([WhatsApp()]).notify(full({"whatsapp"}), "yo"), "muting applies to custom channels too")


def test_rejects_blank_message():
    raises(ValueError, lambda: svc(Outbox()).notify(full(), "  "), "blank message")
