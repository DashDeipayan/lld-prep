// Current production code: works single-threaded, breaks under load.
final class IdGenerator {
    private static IdGenerator instance;
    private long counter = 0;

    private IdGenerator() {
        // simulate expensive initialisation (loading config); keep this
        try { Thread.sleep(20); } catch (InterruptedException e) { Thread.currentThread().interrupt(); }
    }

    static IdGenerator getInstance() {
        if (instance == null) {
            instance = new IdGenerator();
        }
        return instance;
    }

    long nextId() {
        counter++;
        return counter;
    }

    String nextId(String prefix) {
        return prefix + "-" + nextId();
    }
}
