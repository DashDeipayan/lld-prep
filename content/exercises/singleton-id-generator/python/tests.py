import re
import threading

from lldtest import check, eq
from solution import IdGenerator


# Runs first (tests run in the order they are defined), before anything has created the instance.
def test_concurrent_first_access_returns_one_instance():
    n = 32
    start = threading.Barrier(n)
    got = [None] * n

    def worker(i):
        start.wait()
        got[i] = IdGenerator.get_instance()

    threads = [threading.Thread(target=worker, args=(i,)) for i in range(n)]
    for t in threads:
        t.start()
    for t in threads:
        t.join(5)
    eq(1, len({id(g) for g in got}), "all threads must get the same instance")


def test_sequential_ids_increase():
    g = IdGenerator.get_instance()
    a, b, c = g.next_id(), g.next_id(), g.next_id()
    check(b == a + 1 and c == b + 1, "ids should be consecutive: %d, %d, %d" % (a, b, c))


def test_concurrent_ids_are_unique():
    g = IdGenerator.get_instance()
    workers, per = 16, 5000
    start = threading.Barrier(workers)
    results = [[] for _ in range(workers)]

    def worker(i):
        start.wait()
        out = results[i]
        for _ in range(per):
            out.append(g.next_id())

    threads = [threading.Thread(target=worker, args=(i,)) for i in range(workers)]
    for t in threads:
        t.start()
    for t in threads:
        t.join(10)
    seen = set()
    for r in results:
        seen.update(r)
    eq(workers * per, len(seen), "every id must be unique (lost updates mean duplicates)")


def test_prefix_format():
    code = IdGenerator.get_instance().next_id_with_prefix("ORD")
    check(re.fullmatch(r"ORD-\d{6,}", code) is not None, "expected ORD- followed by 6+ digits, got " + code)
