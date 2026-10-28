import java.util.*;

final class KVStore {
    private final Map<String, String> data = new HashMap<>();
    private final Map<String, Integer> counts = new HashMap<>();
    /** One undo log per open block: key -> value before this block first touched it (null = absent). */
    private final Deque<Map<String, Optional<String>>> undo = new ArrayDeque<>();

    synchronized void set(String key, String value) { write(key, Objects.requireNonNull(value)); }

    synchronized Optional<String> get(String key) { return Optional.ofNullable(data.get(key)); }

    synchronized void delete(String key) { if (data.containsKey(key)) write(key, null); }

    synchronized int count(String value) { return counts.getOrDefault(value, 0); }

    synchronized void begin() { undo.push(new HashMap<>()); }

    synchronized void rollback() {
        if (undo.isEmpty()) throw new IllegalStateException("no transaction");
        for (Map.Entry<String, Optional<String>> e : undo.pop().entrySet()) put(e.getKey(), e.getValue().orElse(null));
    }

    synchronized void commit() {
        if (undo.isEmpty()) throw new IllegalStateException("no transaction");
        undo.clear(); // changes already live in `data`
    }

    /** A user-visible change: record the old value in the innermost block (first touch only). */
    private void write(String key, String value) {
        if (!undo.isEmpty()) undo.peek().putIfAbsent(key, Optional.ofNullable(data.get(key)));
        put(key, value);
    }

    /** The single place that keeps data and value counts consistent. */
    private void put(String key, String value) {
        String old = value == null ? data.remove(key) : data.put(key, value);
        if (old != null) counts.merge(old, -1, (a, b) -> a + b == 0 ? null : a + b);
        if (value != null) counts.merge(value, 1, Integer::sum);
    }
}
