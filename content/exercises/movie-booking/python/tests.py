import threading

from lldtest import check, eq, raises
from solution import (Booking, BookingService, BookingStateError, HoldNotActiveError, PaymentFailedError,
                      PaymentGateway, SeatUnavailableError)

MIN = 60_000


class FakeGateway(PaymentGateway):
    def __init__(self):
        self.charges = []
        self.approve = True

    def charge(self, user, amount):
        self.charges.append("%s:%d" % (user, amount))
        return self.approve


class Env:
    def __init__(self):
        self.now = 0
        self.gw = FakeGateway()
        b = self.svc = BookingService(self.gw, lambda: self.now, 5 * MIN)
        b.add_theatre("T1", "Bengaluru")
        b.add_theatre("T2", "Mumbai")
        b.add_screen("T1", "SC1", 2, 1, 3)
        b.add_screen("T2", "SC2", 1, 1, 2)
        b.add_movie("M1", "Dune", 120)
        b.add_movie("M2", "Up", 90)


def test_screen_layout_and_availability():
    b = Env().svc
    s = b.add_show("M1", "SC1", 60 * MIN, 200, 350)
    eq(["A1", "A2", "A3", "B1", "B2", "B3", "C1", "C2", "C3"], b.available_seats(s), "rows A-B regular, C premium")
    raises(ValueError, lambda: b.add_screen("T9", "X", 1, 1, 1), "unknown theatre")


def test_show_overlap_on_same_screen():
    b = Env().svc
    b.add_show("M1", "SC1", 60 * MIN, 200, 350)
    raises(BookingStateError, lambda: b.add_show("M2", "SC1", 179 * MIN, 200, 350), "starts before previous ends")
    eq("S2", b.add_show("M2", "SC1", 180 * MIN, 200, 350), "back-to-back is fine")
    eq("S3", b.add_show("M1", "SC2", 60 * MIN, 200, 350), "same time, different screen")


def test_find_shows_by_city_movie_and_time():
    env = Env()
    b = env.svc
    late = b.add_show("M1", "SC1", 300 * MIN, 200, 350)
    early = b.add_show("M1", "SC1", 60 * MIN, 200, 350)
    b.add_show("M1", "SC2", 60 * MIN, 200, 350)
    b.add_show("M2", "SC1", 500 * MIN, 200, 350)
    eq([early, late], b.find_shows("bengaluru", "M1"), "city match is case-insensitive, sorted by start")
    env.now = 61 * MIN
    eq([late], b.find_shows("Bengaluru", "M1"), "started shows excluded")


def test_hold_confirm_pricing():
    env = Env()
    b = env.svc
    s = b.add_show("M1", "SC1", 60 * MIN, 200, 350)
    h = b.hold(s, "asha", ["C1", "A1"])
    eq("H1", h, "hold id")
    bk = b.confirm(h)
    eq(Booking("B1", s, "asha", ["A1", "C1"], 550, "CONFIRMED"), bk, "regular 200 + premium 350, seats sorted")
    eq(["asha:550"], env.gw.charges, "gateway charged once with the total")
    raises(SeatUnavailableError, lambda: b.hold(s, "ravi", ["A1"]), "booked seat")
    raises(HoldNotActiveError, lambda: b.confirm(h), "hold already used")


def test_payment_failure_releases_seats():
    env = Env()
    b = env.svc
    s = b.add_show("M1", "SC1", 60 * MIN, 200, 350)
    h = b.hold(s, "asha", ["A1", "A2"])
    env.gw.approve = False
    raises(PaymentFailedError, lambda: b.confirm(h), "declined")
    eq(9, len(b.available_seats(s)), "seats released after failed payment")


def test_hold_expiry_and_show_start():
    env = Env()
    b = env.svc
    s = b.add_show("M1", "SC1", 60 * MIN, 200, 350)
    h = b.hold(s, "asha", ["A1"])
    env.now += 5 * MIN
    raises(HoldNotActiveError, lambda: b.confirm(h), "expired hold")
    eq([], env.gw.charges, "no charge for expired hold")
    b.hold(s, "ravi", ["A1"])
    env.now = 60 * MIN
    raises(BookingStateError, lambda: b.hold(s, "x", ["B1"]), "show started")


def test_cancellation():
    env = Env()
    b = env.svc
    s = b.add_show("M1", "SC1", 60 * MIN, 200, 350)
    bk = b.confirm(b.hold(s, "asha", ["B2"]))
    c = b.cancel(bk.id)
    eq("CANCELLED", c.status, "cancelled")
    check("B2" in b.available_seats(s), "seat back on sale")
    raises(BookingStateError, lambda: b.cancel(bk.id), "double cancel")
    bk2 = b.confirm(b.hold(s, "ravi", ["B2"]))
    env.now = 61 * MIN
    raises(BookingStateError, lambda: b.cancel(bk2.id), "after show start")


def test_no_double_booking_under_contention():
    b = Env().svc
    s = b.add_show("M1", "SC1", 60 * MIN, 200, 350)
    start = threading.Barrier(40)
    results = [None] * 40

    def attempt(i):
        want = ["A%d" % (1 + i % 3), "C%d" % (1 + (i // 3) % 3)]
        start.wait()
        try:
            results[i] = b.confirm(b.hold(s, "u%d" % i, want))
        except SeatUnavailableError:
            pass

    threads = [threading.Thread(target=attempt, args=(i,)) for i in range(40)]
    for t in threads:
        t.start()
    for t in threads:
        t.join(10)
    owner = {}
    for bk in results:
        if bk is None:
            continue
        for seat in bk.seats:
            check(owner.setdefault(seat, bk.id) == bk.id, "seat %s sold twice" % seat)
    check(owner, "some bookings succeed")
