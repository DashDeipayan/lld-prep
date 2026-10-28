import java.util.*;

public class Main {
    public static void main(String[] args) {
        long[] now = {0};
        int[] calls = {0};
        Sender flakySms = (to, msg) -> {
            if (++calls[0] % 3 != 0) throw new IllegalStateException("gateway timeout");
            System.out.println("  SMS -> " + to + ": " + msg);
        };
        List<String> log = new ArrayList<>();
        Sender sender = new LoggingSender(
                new RateLimitedSender(new RetryingSender(flakySms, 3), () -> now[0], 2, 3_600_000), log);
        for (int i = 1; i <= 3; i++) {
            try {
                sender.send("+1555", "promo " + i);
            } catch (RuntimeException e) {
                System.out.println("  rejected: " + e.getMessage());
            }
        }
        now[0] += 3_600_000;
        sender.send("+1555", "promo 4");
        log.forEach(System.out::println);
    }
}
