import java.util.*;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.locks.*;

final class BoundedBlockingQueue<T> {
    BoundedBlockingQueue(int capacity) {
        // TODO
    }

    void put(T item) throws InterruptedException { throw new UnsupportedOperationException("TODO"); }

    Optional<T> take() throws InterruptedException { throw new UnsupportedOperationException("TODO"); }

    boolean offer(T item, long timeoutMillis) throws InterruptedException { throw new UnsupportedOperationException("TODO"); }

    Optional<T> poll(long timeoutMillis) throws InterruptedException { throw new UnsupportedOperationException("TODO"); }

    void close() { throw new UnsupportedOperationException("TODO"); }

    int size() { throw new UnsupportedOperationException("TODO"); }
}
