import java.util.*;

// ---- Provided ----
record User(String id, String email, String phone, String deviceToken, Set<String> muted) {}

class Outbox {
    final List<String> sent = new ArrayList<>();
    void record(String channel, String to, String message) { sent.add(channel + ":" + to + ":" + message); }
}

// ---- The code you are replacing (read it, then delete it) ----
class LegacyNotifications {
    // Problems: new channel => edit this method; hard-wired clients; one failure stops the rest.
    int notify(User u, String msg, Outbox outbox) {
        int sent = 0;
        if (u.email() != null && !u.email().isEmpty() && !u.muted().contains("email")) {
            outbox.record("email", u.email(), msg); sent++;
        }
        if (u.phone() != null && !u.phone().isEmpty() && !u.muted().contains("sms")) {
            outbox.record("sms", u.phone(), msg); sent++;      // if this threw, push never runs
        }
        if (u.deviceToken() != null && !u.deviceToken().isEmpty() && !u.muted().contains("push")) {
            outbox.record("push", u.deviceToken(), msg); sent++;
        }
        return sent;
    }
}

// ---- Your design ----
interface Notifier {
    String channel();
    boolean canReach(User u);
    void send(User u, String message);
}

class EmailNotifier implements Notifier {
    EmailNotifier(Outbox outbox) { /* TODO */ }
    public String channel() { throw new UnsupportedOperationException("TODO"); }
    public boolean canReach(User u) { throw new UnsupportedOperationException("TODO"); }
    public void send(User u, String message) { throw new UnsupportedOperationException("TODO"); }
}

class SmsNotifier implements Notifier {
    SmsNotifier(Outbox outbox) { /* TODO */ }
    public String channel() { throw new UnsupportedOperationException("TODO"); }
    public boolean canReach(User u) { throw new UnsupportedOperationException("TODO"); }
    public void send(User u, String message) { throw new UnsupportedOperationException("TODO"); }
}

class PushNotifier implements Notifier {
    PushNotifier(Outbox outbox) { /* TODO */ }
    public String channel() { throw new UnsupportedOperationException("TODO"); }
    public boolean canReach(User u) { throw new UnsupportedOperationException("TODO"); }
    public void send(User u, String message) { throw new UnsupportedOperationException("TODO"); }
}

class NotificationService {
    NotificationService(List<Notifier> notifiers) { /* TODO */ }

    /** @return number of successful sends */
    int notify(User u, String message) {
        throw new UnsupportedOperationException("TODO");
    }
}
