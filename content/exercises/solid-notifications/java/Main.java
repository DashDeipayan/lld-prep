import java.util.*;

public class Main {
    public static void main(String[] args) {
        Outbox outbox = new Outbox();
        NotificationService svc = new NotificationService(List.of(
                new EmailNotifier(outbox), new SmsNotifier(outbox), new PushNotifier(outbox)));
        User alice = new User("u1", "alice@example.com", "+15550100", "tok-a", Set.of("sms"));
        System.out.println("sent = " + svc.notify(alice, "Your order shipped"));
        outbox.sent.forEach(System.out::println);
    }
}
