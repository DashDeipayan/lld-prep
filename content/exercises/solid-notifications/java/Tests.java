import java.util.*;

public class Tests {
    static User full(Set<String> muted) { return new User("u1", "a@x.com", "+1555", "tok", muted); }

    static NotificationService svc(Outbox o) {
        return new NotificationService(List.of(new EmailNotifier(o), new SmsNotifier(o), new PushNotifier(o)));
    }

    static void testChannelsAndReachability() {
        Outbox o = new Outbox();
        T.eq("email", new EmailNotifier(o).channel(), "email channel name");
        T.eq("sms", new SmsNotifier(o).channel(), "sms channel name");
        T.eq("push", new PushNotifier(o).channel(), "push channel name");
        User noPhone = new User("u", "a@x.com", "", "tok", Set.of());
        T.check(new EmailNotifier(o).canReach(noPhone), "email reachable");
        T.check(!new SmsNotifier(o).canReach(noPhone), "empty phone -> sms cannot reach");
        User nulls = new User("u", null, null, null, Set.of());
        T.check(!new PushNotifier(o).canReach(nulls), "null token -> push cannot reach");
    }

    static void testSendsOnAllReachableChannelsInOrder() {
        Outbox o = new Outbox();
        T.eq(3, svc(o).notify(full(Set.of()), "hi"), "three channels");
        T.eq(List.of("email:a@x.com:hi", "sms:+1555:hi", "push:tok:hi"), o.sent, "outbox contents and order");
    }

    static void testRespectsMutedChannels() {
        Outbox o = new Outbox();
        T.eq(2, svc(o).notify(full(Set.of("sms")), "hi"), "sms muted");
        T.eq(List.of("email:a@x.com:hi", "push:tok:hi"), o.sent, "no sms recorded");
    }

    static void testSkipsUnreachable() {
        Outbox o = new Outbox();
        User onlyPush = new User("u", "", null, "tok", Set.of());
        T.eq(1, svc(o).notify(onlyPush, "hi"), "only push reachable");
    }

    static void testFailureIsIsolated() {
        Outbox o = new Outbox();
        Notifier broken = new Notifier() {
            public String channel() { return "slack"; }
            public boolean canReach(User u) { return true; }
            public void send(User u, String m) { throw new IllegalStateException("slack down"); }
        };
        NotificationService s = new NotificationService(List.of(new EmailNotifier(o), broken, new PushNotifier(o)));
        T.eq(2, s.notify(full(Set.of()), "hi"), "broken channel counted as not sent");
        T.eq(List.of("email:a@x.com:hi", "push:tok:hi"), o.sent, "channels after the broken one still ran");
    }

    static void testNewChannelWithoutEditingService() {
        List<String> got = new ArrayList<>();
        Notifier whatsapp = new Notifier() {
            public String channel() { return "whatsapp"; }
            public boolean canReach(User u) { return true; }
            public void send(User u, String m) { got.add(m); }
        };
        T.eq(1, new NotificationService(List.of(whatsapp)).notify(full(Set.of()), "yo"), "custom notifier works");
        T.eq(0, new NotificationService(List.of(whatsapp)).notify(full(Set.of("whatsapp")), "yo"), "muting applies to custom channels too");
    }

    static void testRejectsBlankMessage() {
        T.throwsEx(IllegalArgumentException.class, () -> svc(new Outbox()).notify(full(Set.of()), "  "), "blank message");
    }
}
