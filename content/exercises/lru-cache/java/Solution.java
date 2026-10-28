import java.util.*;
import java.util.function.BiConsumer;

final class LRUCache<K, V> {
    LRUCache(int capacity, BiConsumer<K, V> onEvict) {
        // TODO
    }

    Optional<V> get(K key) { throw new UnsupportedOperationException("TODO"); }

    void put(K key, V value) { throw new UnsupportedOperationException("TODO"); }

    int size() { throw new UnsupportedOperationException("TODO"); }

    /** most recently used first */
    List<K> keys() { throw new UnsupportedOperationException("TODO"); }
}
