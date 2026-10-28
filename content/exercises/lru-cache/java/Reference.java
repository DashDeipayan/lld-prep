import java.util.*;
import java.util.function.BiConsumer;

final class LRUCache<K, V> {
    private static final class Node<K, V> {
        K key;
        V value;
        Node<K, V> prev, next;
        Node(K k, V v) { key = k; value = v; }
    }

    private final int capacity;
    private final BiConsumer<K, V> onEvict;
    private final Map<K, Node<K, V>> map = new HashMap<>();
    private final Node<K, V> head = new Node<>(null, null), tail = new Node<>(null, null); // sentinels

    LRUCache(int capacity, BiConsumer<K, V> onEvict) {
        if (capacity < 1) throw new IllegalArgumentException("capacity must be >= 1");
        this.capacity = capacity;
        this.onEvict = onEvict;
        head.next = tail;
        tail.prev = head;
    }

    synchronized Optional<V> get(K key) {
        Node<K, V> n = map.get(key);
        if (n == null) return Optional.empty();
        moveToFront(n);
        return Optional.ofNullable(n.value);
    }

    void put(K key, V value) {
        Node<K, V> evicted = null;
        synchronized (this) {
            Node<K, V> n = map.get(key);
            if (n != null) {
                n.value = value;
                moveToFront(n);
                return;
            }
            n = new Node<>(key, value);
            map.put(key, n);
            addFront(n);
            if (map.size() > capacity) {
                evicted = tail.prev;
                unlink(evicted);
                map.remove(evicted.key);
            }
        }
        // callback outside the lock: user code must not run while we hold it
        if (evicted != null && onEvict != null) onEvict.accept(evicted.key, evicted.value);
    }

    synchronized int size() { return map.size(); }

    synchronized List<K> keys() {
        List<K> out = new ArrayList<>(map.size());
        for (Node<K, V> n = head.next; n != tail; n = n.next) out.add(n.key);
        return out;
    }

    private void moveToFront(Node<K, V> n) {
        unlink(n);
        addFront(n);
    }

    private void addFront(Node<K, V> n) {
        n.prev = head;
        n.next = head.next;
        head.next.prev = n;
        head.next = n;
    }

    private void unlink(Node<K, V> n) {
        n.prev.next = n.next;
        n.next.prev = n.prev;
    }
}
