import java.util.*;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.locks.*;

final class BoundedBlockingQueue<T> {
    private final Deque<T> items = new ArrayDeque<>();
    private final int capacity;
    private final ReentrantLock lock = new ReentrantLock();
    private final Condition notFull = lock.newCondition();
    private final Condition notEmpty = lock.newCondition();
    private boolean closed;

    BoundedBlockingQueue(int capacity) {
        if (capacity < 1) throw new IllegalArgumentException("capacity must be >= 1");
        this.capacity = capacity;
    }

    void put(T item) throws InterruptedException {
        lock.lockInterruptibly();
        try {
            while (items.size() == capacity && !closed) notFull.await();
            if (closed) throw new IllegalStateException("queue closed");
            enqueue(item);
        } finally {
            lock.unlock();
        }
    }

    boolean offer(T item, long timeoutMillis) throws InterruptedException {
        long nanos = TimeUnit.MILLISECONDS.toNanos(timeoutMillis);
        lock.lockInterruptibly();
        try {
            while (items.size() == capacity && !closed) {
                if (nanos <= 0) return false;
                nanos = notFull.awaitNanos(nanos);
            }
            if (closed) throw new IllegalStateException("queue closed");
            enqueue(item);
            return true;
        } finally {
            lock.unlock();
        }
    }

    Optional<T> take() throws InterruptedException {
        lock.lockInterruptibly();
        try {
            while (items.isEmpty() && !closed) notEmpty.await();
            return items.isEmpty() ? Optional.empty() : Optional.of(dequeue());
        } finally {
            lock.unlock();
        }
    }

    Optional<T> poll(long timeoutMillis) throws InterruptedException {
        long nanos = TimeUnit.MILLISECONDS.toNanos(timeoutMillis);
        lock.lockInterruptibly();
        try {
            while (items.isEmpty() && !closed) {
                if (nanos <= 0) return Optional.empty();
                nanos = notEmpty.awaitNanos(nanos);
            }
            return items.isEmpty() ? Optional.empty() : Optional.of(dequeue());
        } finally {
            lock.unlock();
        }
    }

    void close() {
        lock.lock();
        try {
            closed = true;
            notFull.signalAll();
            notEmpty.signalAll();
        } finally {
            lock.unlock();
        }
    }

    int size() {
        lock.lock();
        try {
            return items.size();
        } finally {
            lock.unlock();
        }
    }

    private void enqueue(T item) {
        items.addLast(item);
        notEmpty.signal();
    }

    private T dequeue() {
        T x = items.removeFirst();
        notFull.signal();
        return x;
    }
}
