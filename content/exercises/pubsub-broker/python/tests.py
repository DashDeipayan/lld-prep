import threading
import time
from collections import Counter

from lldtest import check, eq, raises
from solution import Broker, Message, TopicExistsError, UnknownGroupError, UnknownTopicError


def test_publish_assigns_sequential_offsets():
    b = Broker()
    b.create_topic("t")
    eq(0, b.publish("t", "a"), "first offset")
    eq(1, b.publish("t", "b"), "second offset")
    raises(TopicExistsError, lambda: b.create_topic("t"), "duplicate topic")
    raises(UnknownTopicError, lambda: b.publish("nope", "x"), "unknown topic")


def test_from_beginning_vs_latest():
    b = Broker()
    b.create_topic("t")
    b.publish("t", "old")
    b.subscribe("t", "g1", True)
    b.subscribe("t", "g2", False)
    b.publish("t", "new")
    eq([Message(0, "old"), Message(1, "new")], b.poll("t", "g1", 10), "from beginning")
    eq([Message(1, "new")], b.poll("t", "g2", 10), "latest only")
    eq([], b.poll("t", "g1", 10), "nothing new")


def test_poll_batches_and_lag():
    b = Broker()
    b.create_topic("t")
    b.subscribe("t", "g", True)
    for i in range(5):
        b.publish("t", "m%d" % i)
    eq(5, b.lag("t", "g"), "lag before")
    eq([Message(0, "m0"), Message(1, "m1")], b.poll("t", "g", 2), "batch of 2")
    eq(3, b.lag("t", "g"), "lag after")
    eq(3, len(b.poll("t", "g", 100)), "rest")
    eq(0, b.lag("t", "g"), "caught up")


def test_resubscribe_keeps_offset_and_seek():
    b = Broker()
    b.create_topic("t")
    b.subscribe("t", "g", True)
    b.publish("t", "a")
    b.publish("t", "b")
    b.poll("t", "g", 1)
    b.subscribe("t", "g", True)
    eq([Message(1, "b")], b.poll("t", "g", 10), "re-subscribe is a no-op")
    b.seek("t", "g", 0)
    eq(2, len(b.poll("t", "g", 10)), "replay after seek(0)")
    raises(ValueError, lambda: b.seek("t", "g", 3), "past end")
    raises(UnknownGroupError, lambda: b.poll("t", "unknown", 1), "unknown group")


def test_groups_share_work_exactly_once_under_concurrency():
    b = Broker()
    b.create_topic("events")
    b.subscribe("events", "billing", True)
    b.subscribe("events", "audit", True)
    start = threading.Barrier(8)
    seen = Counter()
    seen_lock = threading.Lock()

    def produce(pid):
        start.wait()
        for i in range(500):
            b.publish("events", "%d:%d" % (pid, i))

    def consume():
        start.wait()
        deadline = time.monotonic() + 5
        while time.monotonic() < deadline:
            with seen_lock:
                if len(seen) >= 2000:
                    return
            for m in b.poll("events", "billing", 10):
                with seen_lock:
                    seen[m.offset] += 1

    threads = [threading.Thread(target=produce, args=(p,)) for p in range(4)]
    threads += [threading.Thread(target=consume) for _ in range(4)]
    for t in threads:
        t.start()
    for t in threads:
        t.join(10)
    eq(2000, len(seen), "every message consumed by the group")
    check(all(n == 1 for n in seen.values()), "no message delivered twice within a group")
    eq(2000, len(b.poll("events", "audit", 5000)), "other group independently gets everything")
