import random
import threading

from lldtest import check, eq, raises
from solution import LRUCache


def test_basic_get_put():
    c = LRUCache(2)
    c.put("a", 1)
    c.put("b", 2)
    eq(1, c.get("a"), "hit")
    eq(None, c.get("z"), "miss")
    eq(-1, c.get("z", -1), "miss with a default")
    eq(2, c.size(), "size")
    raises(ValueError, lambda: LRUCache(0), "capacity >= 1")


def test_evicts_least_recently_used():
    ev = []
    c = LRUCache(2, lambda k, v: ev.append("%s=%s" % (k, v)))
    c.put("a", 1)
    c.put("b", 2)
    c.get("a")
    c.put("c", 3)
    eq(["b=2"], ev, "b was LRU because a was read")
    eq(None, c.get("b"), "b gone")
    eq(["c", "a"], c.keys(), "MRU first")


def test_update_refreshes_and_does_not_evict():
    ev = []
    c = LRUCache(2, lambda k, v: ev.append(k))
    c.put("a", 1)
    c.put("b", 2)
    c.put("a", 10)
    eq([], ev, "update is not an eviction")
    eq(["a", "b"], c.keys(), "update moves to front")
    c.put("c", 3)
    eq(["b"], ev, "b evicted next")
    eq(10, c.get("a"), "updated value")


def test_capacity_one():
    c = LRUCache(1)
    for i in range(5):
        c.put(i, i)
    eq([4], c.keys(), "only the latest survives")


def test_concurrent_access_keeps_invariants():
    c = LRUCache(64)
    start = threading.Barrier(8)
    errors = []

    def worker(seed):
        r = random.Random(seed)
        start.wait()
        for _ in range(10_000):
            k = r.randrange(200)
            if r.random() < 0.5:
                c.put(k, k)
            else:
                v = c.get(k)
                if v is not None and v != k:
                    errors.append("wrong value for %d" % k)

    threads = [threading.Thread(target=worker, args=(t,)) for t in range(8)]
    for t in threads:
        t.start()
    for t in threads:
        t.join(20)
    eq([], errors[:3], "values never mixed up")
    check(c.size() <= 64, "size <= capacity, got %d" % c.size())
    keys = c.keys()
    eq(c.size(), len(keys), "list and map agree")
    eq(len(keys), len(set(keys)), "no duplicate nodes in the list")
