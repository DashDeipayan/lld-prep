import java.util.*;
import java.util.concurrent.*;

public class Tests {
    static void testPublishAssignsSequentialOffsets() {
        Broker b = new Broker();
        b.createTopic("t");
        T.eq(0L, b.publish("t", "a"), "first offset");
        T.eq(1L, b.publish("t", "b"), "second offset");
        T.throwsEx(IllegalStateException.class, () -> b.createTopic("t"), "duplicate topic");
        T.throwsEx(NoSuchElementException.class, () -> b.publish("nope", "x"), "unknown topic");
    }

    static void testFromBeginningVsLatest() {
        Broker b = new Broker();
        b.createTopic("t");
        b.publish("t", "old");
        b.subscribe("t", "g1", true);
        b.subscribe("t", "g2", false);
        b.publish("t", "new");
        T.eq(List.of(new Message(0, "old"), new Message(1, "new")), b.poll("t", "g1", 10), "from beginning");
        T.eq(List.of(new Message(1, "new")), b.poll("t", "g2", 10), "latest only");
        T.eq(List.of(), b.poll("t", "g1", 10), "nothing new");
    }

    static void testPollBatchesAndLag() {
        Broker b = new Broker();
        b.createTopic("t");
        b.subscribe("t", "g", true);
        for (int i = 0; i < 5; i++) b.publish("t", "m" + i);
        T.eq(5L, b.lag("t", "g"), "lag before");
        T.eq(List.of(new Message(0, "m0"), new Message(1, "m1")), b.poll("t", "g", 2), "batch of 2");
        T.eq(3L, b.lag("t", "g"), "lag after");
        T.eq(3, b.poll("t", "g", 100).size(), "rest");
        T.eq(0L, b.lag("t", "g"), "caught up");
    }

    static void testResubscribeKeepsOffsetAndSeek() {
        Broker b = new Broker();
        b.createTopic("t");
        b.subscribe("t", "g", true);
        b.publish("t", "a");
        b.publish("t", "b");
        b.poll("t", "g", 1);
        b.subscribe("t", "g", true);
        T.eq(List.of(new Message(1, "b")), b.poll("t", "g", 10), "re-subscribe is a no-op");
        b.seek("t", "g", 0);
        T.eq(2, b.poll("t", "g", 10).size(), "replay after seek(0)");
        T.throwsEx(IllegalArgumentException.class, () -> b.seek("t", "g", 3), "past end");
        T.throwsEx(IllegalStateException.class, () -> b.poll("t", "unknown", 1), "unknown group");
    }

    static void testGroupsShareWorkExactlyOnceUnderConcurrency() throws Exception {
        Broker b = new Broker();
        b.createTopic("events");
        b.subscribe("events", "billing", true);
        b.subscribe("events", "audit", true);
        ExecutorService pool = Executors.newFixedThreadPool(8);
        CountDownLatch go = new CountDownLatch(1);
        List<Future<?>> producers = new ArrayList<>();
        for (int p = 0; p < 4; p++) {
            int id = p;
            producers.add(pool.submit(() -> { go.await(); for (int i = 0; i < 500; i++) b.publish("events", id + ":" + i); return null; }));
        }
        ConcurrentHashMap<Long, Integer> seen = new ConcurrentHashMap<>();
        List<Future<?>> consumers = new ArrayList<>();
        for (int c = 0; c < 4; c++) consumers.add(pool.submit(() -> {
            go.await();
            long deadline = System.currentTimeMillis() + 5000;
            while (seen.size() < 2000 && System.currentTimeMillis() < deadline) {
                for (Message m : b.poll("events", "billing", 10)) seen.merge(m.offset(), 1, Integer::sum);
            }
            return null;
        }));
        go.countDown();
        for (Future<?> f : producers) f.get(10, TimeUnit.SECONDS);
        for (Future<?> f : consumers) f.get(10, TimeUnit.SECONDS);
        pool.shutdownNow();
        T.eq(2000, seen.size(), "every message consumed by the group");
        T.check(seen.values().stream().allMatch(n -> n == 1), "no message delivered twice within a group");
        T.eq(2000, b.poll("events", "audit", 5000).size(), "other group independently gets everything");
    }
}
