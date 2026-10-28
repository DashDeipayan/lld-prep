import java.util.*;
import java.util.concurrent.*;
import java.util.concurrent.atomic.*;

public class Tests {
    static void testFifoAndSize() throws Exception {
        BoundedBlockingQueue<String> q = new BoundedBlockingQueue<>(3);
        q.put("a"); q.put("b"); q.put("c");
        T.eq(3, q.size(), "size");
        T.eq(Optional.of("a"), q.take(), "fifo 1");
        T.eq(Optional.of("b"), q.take(), "fifo 2");
        T.eq(1, q.size(), "size after takes");
        T.throwsEx(IllegalArgumentException.class, () -> new BoundedBlockingQueue<String>(0), "capacity >= 1");
    }

    static void testTimeoutsWhenFullOrEmpty() throws Exception {
        BoundedBlockingQueue<Integer> q = new BoundedBlockingQueue<>(1);
        T.eq(Optional.empty(), q.poll(50), "poll on empty times out");
        T.check(q.offer(1, 50), "offer with room");
        long t0 = System.nanoTime();
        T.check(!q.offer(2, 80), "offer on full times out");
        T.check(System.nanoTime() - t0 >= TimeUnit.MILLISECONDS.toNanos(60), "offer actually waited");
    }

    static void testPutBlocksUntilTake() throws Exception {
        BoundedBlockingQueue<Integer> q = new BoundedBlockingQueue<>(1);
        q.put(1);
        AtomicBoolean putDone = new AtomicBoolean();
        Thread p = new Thread(() -> { try { q.put(2); putDone.set(true); } catch (Exception ignored) {} });
        p.start();
        Thread.sleep(80);
        T.check(!putDone.get(), "put must block while full");
        T.eq(Optional.of(1), q.take(), "take frees a slot");
        p.join(1000);
        T.check(putDone.get(), "blocked put completes after take");
        T.eq(Optional.of(2), q.take(), "second item");
    }

    static void testProducerConsumerNoLossNoDuplication() throws Exception {
        BoundedBlockingQueue<Integer> q = new BoundedBlockingQueue<>(8);
        int producers = 4, per = 2000;
        ExecutorService pool = Executors.newFixedThreadPool(8);
        ConcurrentLinkedQueue<Integer> got = new ConcurrentLinkedQueue<>();
        List<Future<?>> prods = new ArrayList<>();
        for (int p = 0; p < producers; p++) {
            int base = p * per;
            prods.add(pool.submit(() -> { for (int i = 0; i < per; i++) q.put(base + i); return null; }));
        }
        List<Future<?>> cons = new ArrayList<>();
        for (int c = 0; c < 4; c++) cons.add(pool.submit(() -> {
            Optional<Integer> x;
            while ((x = q.take()).isPresent()) got.add(x.get());
            return null;
        }));
        for (Future<?> f : prods) f.get(10, TimeUnit.SECONDS);
        q.close();
        for (Future<?> f : cons) f.get(10, TimeUnit.SECONDS);
        pool.shutdownNow();
        T.eq(producers * per, got.size(), "every item consumed exactly once (count)");
        T.eq(producers * per, new HashSet<>(got).size(), "no duplicates");
    }

    static void testCloseWakesBlockedCallers() throws Exception {
        BoundedBlockingQueue<Integer> empty = new BoundedBlockingQueue<>(1);
        AtomicReference<Optional<Integer>> taken = new AtomicReference<>();
        Thread c = new Thread(() -> { try { taken.set(empty.take()); } catch (Exception ignored) {} });
        c.start();

        BoundedBlockingQueue<Integer> full = new BoundedBlockingQueue<>(1);
        full.put(1);
        AtomicReference<Throwable> putErr = new AtomicReference<>();
        Thread p = new Thread(() -> { try { full.put(2); } catch (Throwable e) { putErr.set(e); } });
        p.start();

        Thread.sleep(80);
        empty.close();
        full.close();
        c.join(1000);
        p.join(1000);
        T.check(!c.isAlive() && !p.isAlive(), "close must wake blocked take and put");
        T.eq(Optional.empty(), taken.get(), "blocked take returns empty after close");
        T.check(putErr.get() instanceof IllegalStateException, "blocked put fails with IllegalStateException, got " + putErr.get());
    }

    static void testDrainAfterClose() throws Exception {
        BoundedBlockingQueue<Integer> q = new BoundedBlockingQueue<>(3);
        q.put(1); q.put(2);
        q.close();
        q.close();
        T.throwsEx(IllegalStateException.class, () -> { try { q.put(3); } catch (InterruptedException e) { throw new RuntimeException(e); } }, "put after close");
        T.eq(Optional.of(1), q.take(), "remaining items still available");
        T.eq(Optional.of(2), q.poll(10), "poll drains too");
        T.eq(Optional.empty(), q.take(), "then empty, without blocking");
    }
}
