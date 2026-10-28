import java.util.*;
import java.util.concurrent.*;

public class Tests {
    static void testBasicGetPut() {
        LRUCache<String, Integer> c = new LRUCache<>(2, null);
        c.put("a", 1);
        c.put("b", 2);
        T.eq(Optional.of(1), c.get("a"), "hit");
        T.eq(Optional.empty(), c.get("z"), "miss");
        T.eq(2, c.size(), "size");
        T.throwsEx(IllegalArgumentException.class, () -> new LRUCache<String, Integer>(0, null), "capacity >= 1");
    }

    static void testEvictsLeastRecentlyUsed() {
        List<String> ev = new ArrayList<>();
        LRUCache<String, Integer> c = new LRUCache<>(2, (k, v) -> ev.add(k + "=" + v));
        c.put("a", 1);
        c.put("b", 2);
        c.get("a");
        c.put("c", 3);
        T.eq(List.of("b=2"), ev, "b was LRU because a was read");
        T.eq(Optional.empty(), c.get("b"), "b gone");
        T.eq(List.of("c", "a"), c.keys(), "MRU first");
    }

    static void testUpdateRefreshesAndDoesNotEvict() {
        List<String> ev = new ArrayList<>();
        LRUCache<String, Integer> c = new LRUCache<>(2, (k, v) -> ev.add(k));
        c.put("a", 1);
        c.put("b", 2);
        c.put("a", 10);
        T.eq(List.of(), ev, "update is not an eviction");
        T.eq(List.of("a", "b"), c.keys(), "update moves to front");
        c.put("c", 3);
        T.eq(List.of("b"), ev, "b evicted next");
        T.eq(Optional.of(10), c.get("a"), "updated value");
    }

    static void testCapacityOne() {
        LRUCache<Integer, Integer> c = new LRUCache<>(1, null);
        for (int i = 0; i < 5; i++) c.put(i, i);
        T.eq(List.of(4), c.keys(), "only the latest survives");
    }

    static void testConcurrentAccessKeepsInvariants() throws Exception {
        LRUCache<Integer, Integer> c = new LRUCache<>(64, null);
        ExecutorService pool = Executors.newFixedThreadPool(8);
        CountDownLatch go = new CountDownLatch(1);
        List<Future<?>> fs = new ArrayList<>();
        for (int t = 0; t < 8; t++) {
            int seed = t;
            fs.add(pool.submit(() -> {
                go.await();
                Random r = new Random(seed);
                for (int i = 0; i < 20_000; i++) {
                    int k = r.nextInt(200);
                    if (r.nextBoolean()) c.put(k, k); else c.get(k).ifPresent(v -> { if (v != k) throw new AssertionError("wrong value for " + k); });
                }
                return null;
            }));
        }
        go.countDown();
        for (Future<?> f : fs) f.get(20, TimeUnit.SECONDS);
        pool.shutdownNow();
        T.check(c.size() <= 64, "size <= capacity, got " + c.size());
        List<Integer> keys = c.keys();
        T.eq(c.size(), keys.size(), "list and map agree");
        T.eq(keys.size(), new HashSet<>(keys).size(), "no duplicate nodes in the list");
    }
}
