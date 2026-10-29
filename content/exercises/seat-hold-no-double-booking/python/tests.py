import threading

from lldtest import check, eq, raises
from solution import HoldNotActiveError, SeatInventory, SeatUnavailableError


class FakeClock:
    def __init__(self, now=1_000):
        self.now = now

    def __call__(self):
        return self.now


def inv(*ids):
    clock = FakeClock()
    return SeatInventory(list(ids), 1_000, clock), clock


def test_hold_and_status():
    s, _ = inv("A1", "A2", "A3")
    h = s.hold("u", ["A1", "A3"])
    eq("H1", h, "first hold id")
    eq("HELD", s.status("A1"), "held")
    eq(["A2"], s.available(), "available sorted")
    eq("H2", s.hold("v", ["A2"]), "ids increase")


def test_all_or_nothing():
    s, _ = inv("A1", "A2", "A3")
    s.hold("u", ["A2"])
    raises(SeatUnavailableError, lambda: s.hold("v", ["A1", "A2", "A3"]), "one seat taken")
    eq("AVAILABLE", s.status("A1"), "A1 not held by the failed attempt")
    eq("AVAILABLE", s.status("A3"), "A3 not held by the failed attempt")


def test_expiry_is_lazy_and_exact():
    s, clock = inv("A1")
    h = s.hold("u", ["A1"])
    clock.now += 999
    eq("HELD", s.status("A1"), "still held 1ms before expiry")
    clock.now += 1
    eq("AVAILABLE", s.status("A1"), "expired at exactly now == expires_at")
    raises(HoldNotActiveError, lambda: s.confirm(h), "cannot confirm an expired hold")
    eq("H2", s.hold("v", ["A1"]), "seat can be re-held after expiry")


def test_confirm_and_release():
    s, clock = inv("A1", "A2")
    h1 = s.hold("u", ["A1"])
    h2 = s.hold("v", ["A2"])
    s.confirm(h1)
    s.release(h2)
    clock.now += 10_000
    eq("BOOKED", s.status("A1"), "booked seats never expire")
    eq("AVAILABLE", s.status("A2"), "released")
    raises(HoldNotActiveError, lambda: s.confirm(h1), "double confirm")
    raises(HoldNotActiveError, lambda: s.release(h2), "double release")
    raises(HoldNotActiveError, lambda: s.confirm("H99"), "unknown hold")
    raises(SeatUnavailableError, lambda: s.hold("w", ["A1"]), "booked seat cannot be held")


def test_invalid_input():
    s, _ = inv("A1", "A2")
    raises(ValueError, lambda: s.hold("u", []), "empty")
    raises(ValueError, lambda: s.hold("u", ["A1", "A1"]), "duplicate")
    raises(ValueError, lambda: s.hold("u", ["Z9"]), "unknown seat")
    eq(["A1", "A2"], s.available(), "nothing changed")


def test_no_double_booking_under_contention():
    ids = ["S" + chr(ord("a") + i) for i in range(12)]
    s = SeatInventory(ids, 60_000, lambda: 0)
    n = 64
    start = threading.Barrier(n)
    results = [None] * n

    def worker(t):
        want = [ids[t % 12], ids[(t * 5 + 1) % 12], ids[(t * 7 + 3) % 12]]
        uniq = list(dict.fromkeys(want))  # de-duplicate, keep order
        start.wait()
        try:
            results[t] = (s.hold("u", uniq), uniq)
        except SeatUnavailableError:
            pass

    threads = [threading.Thread(target=worker, args=(t,)) for t in range(n)]
    for t in threads:
        t.start()
    for t in threads:
        t.join(10)
    owner = {}
    ok = 0
    for r in results:
        if r is None:
            continue
        ok += 1
        hold_id, seats = r
        for seat in seats:
            prev = owner.setdefault(seat, hold_id)
            check(prev == hold_id, "seat %s held by both %s and %s" % (seat, prev, hold_id))
            eq("HELD", s.status(seat), "successful hold's seat is HELD")
    check(ok >= 1, "at least one hold must succeed")
