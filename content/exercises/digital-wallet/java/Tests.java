import java.util.*;
import java.util.concurrent.*;

public class Tests {
    static Wallet wallet(String... ids) {
        Wallet w = new Wallet();
        for (String id : ids) w.openAccount(id);
        return w;
    }

    static void testDepositTransferAndHistory() {
        Wallet w = wallet("a", "b");
        T.eq("T1", w.deposit("a", 500), "deposit id");
        T.eq("T2", w.transfer("k1", "a", "b", 200), "transfer id");
        T.eq(300L, w.balance("a"), "a debited");
        T.eq(200L, w.balance("b"), "b credited");
        T.eq(List.of(new Txn("T1", "EXTERNAL", "a", 500), new Txn("T2", "a", "b", 200)), w.history("a"), "a history");
        T.eq(List.of(new Txn("T2", "a", "b", 200)), w.history("b"), "b history");
    }

    static void testValidation() {
        Wallet w = wallet("a", "b");
        w.deposit("a", 100);
        T.throwsEx(IllegalArgumentException.class, () -> w.openAccount("a"), "duplicate account");
        T.throwsEx(IllegalArgumentException.class, () -> w.deposit("a", 0), "zero deposit");
        T.throwsEx(IllegalArgumentException.class, () -> w.transfer("k1", "a", "zz", 10), "unknown account");
        T.throwsEx(IllegalArgumentException.class, () -> w.transfer("k2", "a", "a", 10), "self transfer");
        T.throwsEx(IllegalArgumentException.class, () -> w.transfer("k3", "a", "b", -1), "negative");
        T.throwsEx(InsufficientFundsException.class, () -> w.transfer("k4", "a", "b", 101), "insufficient");
        T.eq(100L, w.balance("a"), "nothing moved");
        T.eq(1, w.history("a").size(), "failed transfers not recorded");
    }

    static void testIdempotentRetries() {
        Wallet w = wallet("a", "b");
        w.deposit("a", 100);
        String t = w.transfer("pay-1", "a", "b", 60);
        T.eq(t, w.transfer("pay-1", "a", "b", 60), "same key -> same txn id");
        T.eq(40L, w.balance("a"), "charged once");
        T.throwsEx(InsufficientFundsException.class, () -> w.transfer("pay-2", "a", "b", 60), "second payment fails");
        w.deposit("a", 1000);
        T.throwsEx(InsufficientFundsException.class, () -> w.transfer("pay-2", "a", "b", 60), "retry of a failed key returns the same failure");
        T.eq(1040L, w.balance("a"), "failed key not re-executed");
    }

    static void testConcurrentRandomTransfersConserveMoney() throws Exception {
        String[] ids = new String[10];
        Wallet w = new Wallet();
        for (int i = 0; i < 10; i++) { ids[i] = "acc" + i; w.openAccount(ids[i]); w.deposit(ids[i], 1_000); }
        ExecutorService pool = Executors.newFixedThreadPool(8);
        List<Future<?>> fs = new ArrayList<>();
        for (int t = 0; t < 8; t++) {
            int seed = t;
            fs.add(pool.submit(() -> {
                Random r = new Random(seed);
                for (int i = 0; i < 125; i++) {
                    int a = r.nextInt(10), b = (a + 1 + r.nextInt(9)) % 10;
                    try { w.transfer("t" + seed + "-" + i, ids[a], ids[b], 1 + r.nextInt(300)); } catch (InsufficientFundsException ignored) {}
                }
                return null;
            }));
        }
        for (Future<?> f : fs) f.get(10, TimeUnit.SECONDS);
        pool.shutdownNow();
        long total = 0;
        for (String id : ids) {
            long b = w.balance(id);
            T.check(b >= 0, id + " went negative: " + b);
            total += b;
        }
        T.eq(10_000L, total, "money conserved");
    }

    static void testOppositeDirectionTransfersDoNotDeadlock() throws Exception {
        Wallet w = wallet("A", "B");
        w.deposit("A", 1_000_000);
        w.deposit("B", 1_000_000);
        ExecutorService pool = Executors.newFixedThreadPool(2);
        CountDownLatch go = new CountDownLatch(1);
        Future<?> ab = pool.submit(() -> { go.await(); for (int i = 0; i < 200; i++) w.transfer("ab" + i, "A", "B", 1); return null; });
        Future<?> ba = pool.submit(() -> { go.await(); for (int i = 0; i < 200; i++) w.transfer("ba" + i, "B", "A", 1); return null; });
        go.countDown();
        ab.get(6, TimeUnit.SECONDS);
        ba.get(6, TimeUnit.SECONDS);
        pool.shutdownNow();
        T.eq(2_000_000L, w.balance("A") + w.balance("B"), "conserved");
    }

    static void testConcurrentSameKeyExecutesOnce() throws Exception {
        Wallet w = wallet("a", "b");
        w.deposit("a", 1_000);
        ExecutorService pool = Executors.newFixedThreadPool(16);
        CountDownLatch go = new CountDownLatch(1);
        List<Future<String>> fs = new ArrayList<>();
        for (int i = 0; i < 50; i++) fs.add(pool.submit(() -> { go.await(); return w.transfer("same", "a", "b", 10); }));
        go.countDown();
        Set<String> ids = new HashSet<>();
        for (Future<String> f : fs) ids.add(f.get(5, TimeUnit.SECONDS));
        pool.shutdownNow();
        T.eq(1, ids.size(), "every retry sees the same txn id");
        T.eq(990L, w.balance("a"), "money moved exactly once");
    }
}
