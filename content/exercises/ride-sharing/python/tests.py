import threading

from lldtest import eq, raises
from solution import (DriverInfo, Location, NearestDriverStrategy, NoDriverAvailableError, Rate, RideService,
                      RideStateError, StandardPricing, TripStatus, VehicleClass)

MINI, SEDAN, SUV = VehicleClass.MINI, VehicleClass.SEDAN, VehicleClass.SUV
RATES = {MINI: Rate(50, 10, 2), SEDAN: Rate(80, 15, 3), SUV: Rate(100, 20, 5)}


class Env:
    def __init__(self):
        self.now = 0
        s = self.svc = RideService(NearestDriverStrategy(10), StandardPricing(RATES, 100), lambda: self.now)
        s.add_driver("d1", MINI, Location(0, 0), 4.5)
        s.add_driver("d2", MINI, Location(5, 5), 4.9)
        s.add_driver("d3", SEDAN, Location(1, 1), 4.7)
        for d in ("d1", "d2", "d3"):
            s.set_online(d, True)


def test_nearest_strategy_tie_breaks():
    m = NearestDriverStrategy(5)
    p = Location(0, 0)
    a = DriverInfo("a", MINI, Location(2, 1), 4.1)
    b = DriverInfo("b", MINI, Location(1, 2), 4.8)
    c = DriverInfo("c", MINI, Location(3, 0), 4.8)
    z = DriverInfo("z", MINI, Location(0, 9), 5.0)
    eq("b", m.match(p, [a, b, c, z]), "distance 3 tie: higher rating wins")
    eq("b", m.match(p, [c, b]), "same distance and rating: lower id wins")
    eq(None, m.match(p, [z]), "out of range")


def test_pricing():
    p = StandardPricing(RATES, 100)
    eq(50 + 10 * 7 + 2 * 3, p.fare(MINI, 7, 2 * 60_000 + 1), "minutes rounded up")
    eq((80 + 15 * 4 + 3 * 10) * 150 // 100, StandardPricing(RATES, 150).fare(SEDAN, 4, 10 * 60_000), "1.5x surge")


def test_request_matches_nearest_available_of_class():
    s = Env().svc
    r1 = s.request_ride("asha", Location(1, 0), Location(9, 9), MINI)
    eq("R1", r1, "trip id")
    eq("d1", s.driver_of(r1), "nearest mini")
    eq(TripStatus.ASSIGNED, s.status(r1), "assigned")
    r2 = s.request_ride("ravi", Location(1, 0), Location(9, 9), MINI)
    eq("d2", s.driver_of(r2), "d1 is busy, d2 next")
    raises(NoDriverAvailableError, lambda: s.request_ride("kim", Location(0, 0), Location(1, 1), MINI), "no minis left")
    raises(NoDriverAvailableError, lambda: s.request_ride("kim", Location(0, 0), Location(1, 1), SUV), "no SUVs at all")
    raises(RideStateError, lambda: s.request_ride("asha", Location(0, 0), Location(1, 1), SEDAN), "rider already on a trip")


def test_offline_drivers_not_matched():
    s = Env().svc
    s.set_online("d1", False)
    eq("d2", s.driver_of(s.request_ride("asha", Location(0, 0), Location(3, 3), MINI)), "offline d1 skipped")
    raises(RideStateError, lambda: s.set_online("d2", False), "can't go offline mid-trip")


def test_lifecycle_and_fare():
    env = Env()
    s = env.svc
    r = s.request_ride("asha", Location(0, 1), Location(4, 4), MINI)
    raises(RideStateError, lambda: s.start(r), "must arrive first")
    s.arrived(r)
    env.now = 1_000
    s.start(r)
    env.now = 1_000 + 9 * 60_000 + 30_000
    eq(50 + 10 * 7 + 2 * 10, s.end(r), "fare: distance 7, 9.5 min -> 10")
    eq(TripStatus.COMPLETED, s.status(r), "completed")
    raises(RideStateError, lambda: s.cancel(r), "can't cancel a completed trip")
    r2 = s.request_ride("ravi", Location(4, 4), Location(0, 0), MINI)
    eq("d1", s.driver_of(r2), "d1 free again and now located at (4,4)")


def test_cancellation_frees_driver_and_rider():
    s = Env().svc
    r = s.request_ride("asha", Location(0, 0), Location(2, 2), SEDAN)
    s.arrived(r)
    s.cancel(r)
    eq(TripStatus.CANCELLED, s.status(r), "cancelled")
    r2 = s.request_ride("asha", Location(0, 0), Location(2, 2), SEDAN)
    eq("d3", s.driver_of(r2), "driver and rider both free again")
    s.arrived(r2)
    s.start(r2)
    raises(RideStateError, lambda: s.cancel(r2), "can't cancel in progress")


def test_no_double_assignment_under_contention():
    s = RideService(NearestDriverStrategy(100), StandardPricing(RATES, 100), lambda: 0)
    for i in range(10):
        s.add_driver("d%d" % i, MINI, Location(i, 0), 4.0)
        s.set_online("d%d" % i, True)
    start = threading.Barrier(50)
    assigned = []
    assigned_lock = threading.Lock()

    def rider(i):
        start.wait()
        try:
            d = s.driver_of(s.request_ride("u%d" % i, Location(0, 0), Location(5, 5), MINI))
            with assigned_lock:
                assigned.append(d)
        except NoDriverAvailableError:
            pass

    threads = [threading.Thread(target=rider, args=(i,)) for i in range(50)]
    for t in threads:
        t.start()
    for t in threads:
        t.join(10)
    eq(10, len(assigned), "exactly 10 trips")
    eq(10, len(set(assigned)), "10 distinct drivers")
