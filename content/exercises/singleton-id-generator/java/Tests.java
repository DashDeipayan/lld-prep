import java.util.*;
import java.util.concurrent.*;

public class Tests {
    // Named "testA..." so it runs first, before anything has created the instance.
    static void testAConcurrentFirstAccessReturnsOneInstance() throws Exception {
        int n = 32;
        CountDownLatch go = new CountDownLatch(1);
        ExecutorService pool = Executors.newFixedThreadPool(n);
        List<Future<IdGenerator>> fs = new ArrayList<>();
        for (int i = 0; i < n; i++) fs.add(pool.submit(() -> { go.await(); return IdGenerator.getInstance(); }));
        go.countDown();
        Set<IdGenerator> seen = Collections.newSetFromMap(new IdentityHashMap<>());
        for (Future<IdGenerator> f : fs) seen.add(f.get(5, TimeUnit.SECONDS));
        pool.shutdownNow();
        T.eq(1, seen.size(), "all threads must get the same instance");
    }

    static void testBSequentialIdsIncrease() {
        IdGenerator g = IdGenerator.getInstance();
        long a = g.nextId(), b = g.nextId(), c = g.nextId();
        T.check(b == a + 1 && c == b + 1, "ids should be consecutive: " + a + ", " + b + ", " + c);
    }

    static void testCConcurrentIdsAreUnique() throws Exception {
        IdGenerator g = IdGenerator.getInstance();
        int workers = 16, per = 5000;
        ConcurrentHashMap<Long, Boolean> seen = new ConcurrentHashMap<>();
        CountDownLatch go = new CountDownLatch(1);
        ExecutorService pool = Executors.newFixedThreadPool(workers);
        List<Future<?>> fs = new ArrayList<>();
        for (int w = 0; w < workers; w++) fs.add(pool.submit(() -> {
            go.await();
            for (int i = 0; i < per; i++) seen.put(g.nextId(), true);
            return null;
        }));
        go.countDown();
        for (Future<?> f : fs) f.get(10, TimeUnit.SECONDS);
        pool.shutdownNow();
        T.eq(workers * per, seen.size(), "every id must be unique (lost updates mean duplicates)");
    }

    static void testDPrefixFormat() {
        String id = IdGenerator.getInstance().nextId("ORD");
        T.check(id.matches("ORD-\\d{6,}"), "expected ORD- followed by 6+ digits, got " + id);
    }
}
