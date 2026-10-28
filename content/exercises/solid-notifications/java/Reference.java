import java.util.*;

// ---- Provided ----
record User(String id, String email, String phone, String deviceToken, Set<String> muted) {}

class Outbox {
    final List<String> sent = new ArrayList<>();
    void record(String channel, String to, String message) { sent.add(channel + ":" + to + ":" + message); }
}

interface Notifier {
    String channel();
    boolean canReach(User u);
    void send(User u, String message);
}

/** Shared plumbing: each concrete notifier only says which address it uses. */
abstract class OutboxNotifier implements Notifier {
    private final Outbox outbox;

    OutboxNotifier(Outbox outbox) { this.outbox = Objects.requireNonNull(outbox); }

    protected abstract String address(User u);

    public boolean canReach(User u) {
        String a = address(u);
        return a != null && !a.isBlank();
    }

    public void send(User u, String message) { outbox.record(channel(), address(u), message); }
}

class EmailNotifier extends OutboxNotifier {
    EmailNotifier(Outbox outbox) { super(outbox); }
    public String channel() { return "email"; }
    protected String address(User u) { return u.email(); }
}

class SmsNotifier extends OutboxNotifier {
    SmsNotifier(Outbox outbox) { super(outbox); }
    public String channel() { return "sms"; }
    protected String address(User u) { return u.phone(); }
}

class PushNotifier extends OutboxNotifier {
    PushNotifier(Outbox outbox) { super(outbox); }
    public String channel() { return "push"; }
    protected String address(User u) { return u.deviceToken(); }
}

class NotificationService {
    private final List<Notifier> notifiers;

    NotificationService(List<Notifier> notifiers) { this.notifiers = List.copyOf(notifiers); }

    int notify(User u, String message) {
        if (message == null || message.isBlank()) throw new IllegalArgumentException("message must not be blank");
        int sent = 0;
        for (Notifier n : notifiers) {
            if (u.muted().contains(n.channel()) || !n.canReach(u)) continue;
            try {
                n.send(u, message);
                sent++;
            } catch (RuntimeException e) {
                // isolate the failure; a real system would log + retry/dead-letter here
            }
        }
        return sent;
    }
}
