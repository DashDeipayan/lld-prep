import java.util.*;
import java.util.concurrent.*;

public class Tests {
    static void testFormatters() {
        T.eq("WARN [orders] payment slow", new SimpleFormatter().format("orders", Level.WARN, "payment slow"), "simple");
        T.eq("{\"level\":\"INFO\",\"logger\":\"app\",\"msg\":\"say \\\"hi\\\" \\\\o/\"}",
                new JsonFormatter().format("app", Level.INFO, "say \"hi\" \\o/"), "json with escaping");
    }

    static void testLoggerAndHandlerLevels() {
        List<String> console = new ArrayList<>(), alerts = new ArrayList<>();
        Logger log = new Logger("orders", Level.INFO)
                .addHandler(new MemoryHandler(Level.DEBUG, console, new SimpleFormatter()))
                .addHandler(new MemoryHandler(Level.ERROR, alerts, new SimpleFormatter()));
        log.debug("d");
        log.info("i");
        log.warn("w");
        log.error("e");
        T.eq(List.of("INFO [orders] i", "WARN [orders] w", "ERROR [orders] e"), console, "logger drops DEBUG");
        T.eq(List.of("ERROR [orders] e"), alerts, "handler keeps only ERROR");
    }

    static void testSetLevelAtRuntime() {
        List<String> sink = new ArrayList<>();
        Logger log = new Logger("x", Level.ERROR).addHandler(new MemoryHandler(Level.DEBUG, sink, new SimpleFormatter()));
        log.info("before");
        log.setLevel(Level.DEBUG);
        log.debug("after");
        T.eq(List.of("DEBUG [x] after"), sink, "level change applies immediately");
    }

    static void testOrderAndFailureIsolation() {
        List<String> order = new ArrayList<>();
        Logger log = new Logger("x", Level.DEBUG)
                .addHandler((l, lv, m) -> order.add("first"))
                .addHandler((l, lv, m) -> { throw new IllegalStateException("disk full"); })
                .addHandler((l, lv, m) -> order.add("third"));
        log.info("m");
        T.eq(List.of("first", "third"), order, "handlers in order, failure isolated");
    }

    static void testConcurrentLogging() throws Exception {
        List<String> sink = new ArrayList<>();
        Logger log = new Logger("x", Level.DEBUG).addHandler(new MemoryHandler(Level.DEBUG, sink, new SimpleFormatter()));
        ExecutorService pool = Executors.newFixedThreadPool(8);
        CountDownLatch go = new CountDownLatch(1);
        List<Future<?>> fs = new ArrayList<>();
        for (int t = 0; t < 8; t++) {
            final int id = t;
            fs.add(pool.submit(() -> { go.await(); for (int i = 0; i < 500; i++) log.info(id + "-" + i); return null; }));
        }
        go.countDown();
        for (Future<?> f : fs) f.get(10, TimeUnit.SECONDS);
        pool.shutdownNow();
        T.eq(4000, sink.size(), "no lost lines under concurrency");
    }
}
