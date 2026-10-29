import threading

from lldtest import eq, raises
from solution import CarType, RentalService, Reservation, ReservationError


def svc():
    r = RentalService()
    r.add_branch("B1", "Bengaluru")
    r.add_branch("B2", "Pune")
    r.add_car("B1", "C1", CarType.SEDAN, 2000)
    r.add_car("B1", "C2", CarType.SEDAN, 1500)
    r.add_car("B1", "C3", CarType.SUV, 3000)
    r.add_car("B2", "C4", CarType.SEDAN, 1000)
    return r


def test_search_filters_and_sorts():
    r = svc()
    eq(["C2", "C1"], r.search("bengaluru", CarType.SEDAN, 1, 3), "city + type, sorted by rate")
    eq(["C3"], r.search("Bengaluru", CarType.SUV, 1, 3), "SUV")
    eq([], r.search("Delhi", CarType.SEDAN, 1, 3), "no branch")


def test_reserve_cost_and_overlap():
    r = svc()
    a = r.reserve("asha", "C2", 10, 13)
    eq(Reservation("RS1", "asha", "C2", 10, 13, 4500, "RESERVED"), a, "3 days x 1500")
    raises(ReservationError, lambda: r.reserve("ravi", "C2", 12, 15), "overlap")
    eq("RS2", r.reserve("ravi", "C2", 13, 15).id, "back-to-back ok")
    eq(["C1"], r.search("Bengaluru", CarType.SEDAN, 11, 12), "reserved car hidden from search")
    raises(ValueError, lambda: r.reserve("x", "C2", 5, 5), "empty range")
    raises(ValueError, lambda: r.reserve("x", "ZZ", 1, 2), "unknown car")


def test_cancel():
    r = svc()
    a = r.reserve("asha", "C1", 1, 4)
    eq("CANCELLED", r.cancel(a.id).status, "cancelled")
    eq(["C2", "C1"], r.search("Bengaluru", CarType.SEDAN, 1, 4), "free again")
    raises(ReservationError, lambda: r.cancel(a.id), "double cancel")


def test_return_with_late_fee():
    r = svc()
    a = r.reserve("asha", "C1", 1, 4)
    eq(6000 + 2 * 2000 * 3 // 2, r.return_car(a.id, 6), "2 days late at 1.5x")
    b = r.reserve("ravi", "C1", 6, 8)
    eq(4000, r.return_car(b.id, 7), "early return: no refund")
    raises(ReservationError, lambda: r.return_car(b.id, 9), "already completed")
    raises(ReservationError, lambda: r.cancel(b.id), "can't cancel completed")


def test_single_winner_under_contention():
    r = svc()
    start = threading.Barrier(30)
    ok = []

    def attempt(i):
        s = 10 + i % 3
        start.wait()
        try:
            r.reserve("u", "C3", s, s + 3)
            ok.append(i)
        except ReservationError:
            pass

    threads = [threading.Thread(target=attempt, args=(i,)) for i in range(30)]
    for t in threads:
        t.start()
    for t in threads:
        t.join(10)
    eq(1, len(ok), "overlapping requests: exactly one wins")
