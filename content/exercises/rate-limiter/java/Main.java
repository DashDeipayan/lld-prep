import java.util.*;

public class Main {
    public static void main(String[] args) {
        long[] now = {0};
        Map<String, RateLimiter> limiters = new LinkedHashMap<>();
        limiters.put("token-bucket(3, 2/s)", new TokenBucketLimiter(3, 2, () -> now[0]));
        limiters.put("fixed-window(3/s)", new FixedWindowLimiter(3, 1000, () -> now[0]));
        limiters.put("sliding-log(3/s)", new SlidingWindowLimiter(3, 1000, () -> now[0]));
        long[] times = {0, 100, 200, 300, 600, 1000, 1100, 1250};
        for (var e : limiters.entrySet()) {
            StringBuilder sb = new StringBuilder(String.format("%-22s", e.getKey()));
            for (long t : times) { now[0] = t; sb.append(t).append(e.getValue().allow("user-1") ? "✓ " : "✗ "); }
            System.out.println(sb);
        }
    }
}
