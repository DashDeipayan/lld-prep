import threading
import time

from lldtest import check, eq, raises
from solution import BoundedBlockingQueue, QueueClosedError


def test_fifo_and_size():
    q = BoundedBlockingQueue(3)
    q.put("a")
    q.put("b")
    q.put("c")
    eq(3, q.size(), "size")
    eq("a", q.take(), "fifo 1")
    eq("b", q.take(), "fifo 2")
    eq(1, q.size(), "size after takes")
    raises(ValueError, lambda: BoundedBlockingQueue(0), "capacity >= 1")


def test_timeouts_when_full_or_empty():
    q = BoundedBlockingQueue(1)
    eq(None, q.poll(0.05), "poll on empty times out")
    check(q.offer(1, 0.05), "offer with room")
    t0 = time.monotonic()
    check(not q.offer(2, 0.08), "offer on full times out")
    check(time.monotonic() - t0 >= 0.06, "offer actually waited")


def test_put_blocks_until_take():
    q = BoundedBlockingQueue(1)
    q.put(1)
    done = threading.Event()

    def producer():
        q.put(2)
        done.set()

    threading.Thread(target=producer, daemon=True).start()
    time.sleep(0.08)
    check(not done.is_set(), "put must block while full")
    eq(1, q.take(), "take frees a slot")
    check(done.wait(1), "blocked put completes after take")
    eq(2, q.take(), "second item")


def test_producer_consumer_no_loss_no_duplication():
    q = BoundedBlockingQueue(8)
    producers, per = 4, 2000
    got = []
    got_lock = threading.Lock()

    def produce(base):
        for i in range(per):
            q.put(base + i)

    def consume():
        while True:
            x = q.take()
            if x is None:
                return
            with got_lock:
                got.append(x)

    prods = [threading.Thread(target=produce, args=(p * per,)) for p in range(producers)]
    cons = [threading.Thread(target=consume) for _ in range(4)]
    for t in prods + cons:
        t.start()
    for t in prods:
        t.join(10)
    q.close()
    for t in cons:
        t.join(10)
    eq(producers * per, len(got), "every item consumed exactly once (count)")
    eq(producers * per, len(set(got)), "no duplicates")


def test_close_wakes_blocked_callers():
    empty = BoundedBlockingQueue(1)
    taken = {}
    c = threading.Thread(target=lambda: taken.setdefault("v", empty.take()), daemon=True)
    c.start()

    full = BoundedBlockingQueue(1)
    full.put(1)
    put_err = {}

    def blocked_put():
        try:
            full.put(2)
        except BaseException as e:
            put_err["e"] = e

    p = threading.Thread(target=blocked_put, daemon=True)
    p.start()

    time.sleep(0.08)
    empty.close()
    full.close()
    c.join(1)
    p.join(1)
    check(not c.is_alive() and not p.is_alive(), "close must wake blocked take and put")
    eq(None, taken.get("v", "never returned"), "blocked take returns None after close")
    check(isinstance(put_err.get("e"), QueueClosedError), "blocked put fails with QueueClosedError, got %r" % put_err.get("e"))


def test_drain_after_close():
    q = BoundedBlockingQueue(3)
    q.put(1)
    q.put(2)
    q.close()
    q.close()
    raises(QueueClosedError, lambda: q.put(3), "put after close")
    eq(1, q.take(), "remaining items still available")
    eq(2, q.poll(0.01), "poll drains too")
    eq(None, q.take(), "then None, without blocking")
