import java.util.concurrent.atomic.AtomicLong;

final class IdGenerator {
    private final AtomicLong counter = new AtomicLong();

    private IdGenerator() {
        try { Thread.sleep(20); } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
    }

    // Initialization-on-demand holder: the JVM guarantees class init happens once, lazily, safely.
    private static final class Holder {
        static final IdGenerator INSTANCE = new IdGenerator();
    }

    static IdGenerator getInstance() { return Holder.INSTANCE; }

    long nextId() { return counter.incrementAndGet(); }

    String nextId(String prefix) { return String.format("%s-%06d", prefix, nextId()); }
}
