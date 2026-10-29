import threading

from lldtest import check, eq, raises
from solution import (DisplayBoard, HourlyPricing, NearestSpotSelector, ParkingFullError, ParkingListener, ParkingLot,
                      ParkingStateError, Receipt, Spot, SpotSelector, SpotSize, Vehicle, VehicleType)

RATES = {VehicleType.BIKE: 10, VehicleType.CAR: 20, VehicleType.TRUCK: 50}
MIN = 60_000


def bike(p):
    return Vehicle(p, VehicleType.BIKE)


def car(p):
    return Vehicle(p, VehicleType.CAR)


def truck(p):
    return Vehicle(p, VehicleType.TRUCK)


class FakeClock:
    def __init__(self):
        self.now = 0

    def __call__(self):
        return self.now


def lot(clock=None):
    l = ParkingLot(HourlyPricing(RATES), NearestSpotSelector(), clock or FakeClock())
    l.add_floor(1, 2, 3, 1)
    l.add_floor(2, 1, 1, 1)
    return l


class Recorder(ParkingListener):
    def __init__(self, fmt, log):
        self.fmt, self.log = fmt, log

    def on_event(self, event):
        self.log.append(self.fmt(event))


# ---------------- step 1 ----------------
def s1_fitting_rules():
    small, medium, large = Spot("x", 1, SpotSize.SMALL), Spot("y", 1, SpotSize.MEDIUM), Spot("z", 1, SpotSize.LARGE)
    check(small.can_fit(bike("b")) and medium.can_fit(bike("b")) and large.can_fit(bike("b")), "bike fits anywhere")
    check(not small.can_fit(car("c")) and medium.can_fit(car("c")) and large.can_fit(car("c")), "car needs medium+")
    check(not small.can_fit(truck("t")) and not medium.can_fit(truck("t")) and large.can_fit(truck("t")), "truck needs large")
    eq("y", medium.id, "id")
    eq(SpotSize.MEDIUM, medium.size, "size")


def s1_occupy_and_vacate():
    s = Spot("F1-M1", 1, SpotSize.MEDIUM)
    check(s.is_free(), "new spot is free")
    s.occupy(car("KA1"))
    check(not s.is_free(), "occupied")
    raises(ParkingStateError, lambda: s.occupy(bike("KA2")), "double occupy")
    eq("KA1", s.vacate().plate, "vacate returns the vehicle")
    check(s.is_free(), "free again")
    raises(ParkingStateError, lambda: Spot("z", 1, SpotSize.SMALL).occupy(truck("T")), "occupy with non-fitting vehicle")


# ---------------- step 2 ----------------
def s2_floors_and_availability():
    l = lot()
    eq(3, l.available(SpotSize.SMALL), "small across floors")
    eq(4, l.available(SpotSize.MEDIUM), "medium across floors")
    eq(2, l.available(SpotSize.LARGE), "large across floors")
    raises(ValueError, lambda: l.add_floor(1, 1, 1, 1), "duplicate floor")


# ---------------- step 3 ----------------
def s3_park_issues_tickets_nearest_first():
    clock = FakeClock()
    l = lot(clock)
    clock.now = 5_000
    t1 = l.park(bike("B1"))
    eq("T1", t1.id, "ticket id")
    eq("F1-S1", t1.spot_id, "bike takes the first small spot")
    eq(5_000, t1.entry_ms, "entry time from clock")
    eq("F1-M1", l.park(car("C1")).spot_id, "car takes first medium")
    eq("F1-L1", l.park(truck("TR1")).spot_id, "truck takes first large")
    eq("F2-L1", l.park(truck("TR2")).spot_id, "second truck goes to floor 2")
    eq(3, l.available(SpotSize.MEDIUM), "availability updated")


def s3_full_and_duplicate():
    l = lot()
    l.park(truck("TR1"))
    l.park(truck("TR2"))
    raises(ParkingFullError, lambda: l.park(truck("TR3")), "no large spot left")
    raises(ParkingStateError, lambda: l.park(truck("TR1")), "plate already parked")


def s3_custom_selector_strategy():
    class TopFloorFirst(SpotSelector):
        def select(self, candidates, vehicle):
            return max(candidates, key=lambda s: s.floor, default=None)

    l = ParkingLot(HourlyPricing(RATES), TopFloorFirst(), FakeClock())
    l.add_floor(1, 1, 1, 1)
    l.add_floor(2, 1, 1, 1)
    eq("F2-M1", l.park(car("C1")).spot_id, "custom strategy used without changing ParkingLot")


# ---------------- step 4 ----------------
def s4_hourly_pricing():
    p = HourlyPricing(RATES)
    eq(20, p.fee(VehicleType.CAR, 0), "minimum one hour")
    eq(20, p.fee(VehicleType.CAR, 59 * MIN), "59 min")
    eq(20, p.fee(VehicleType.CAR, 60 * MIN), "exactly 60 min")
    eq(40, p.fee(VehicleType.CAR, 61 * MIN), "61 min -> 2 hours")
    eq(150, p.fee(VehicleType.TRUCK, 180 * MIN), "3h truck")


def s4_unpark_returns_receipt():
    clock = FakeClock()
    l = lot(clock)
    t = l.park(car("C1"))
    clock.now += 90 * MIN
    eq(Receipt("T1", "C1", "F1-M1", 90, 40), l.unpark(t.id), "receipt")
    eq(4, l.available(SpotSize.MEDIUM), "spot freed")
    raises(ValueError, lambda: l.unpark(t.id), "ticket can't be reused")
    raises(ValueError, lambda: l.unpark("T999"), "unknown ticket")
    eq("F1-M1", l.park(car("C1")).spot_id, "same plate can park again")


# ---------------- step 5 ----------------
def s5_events_in_order():
    l = lot()
    log = []
    l.add_listener(Recorder(lambda e: "a:%s:%s" % (e.kind, e.spot_id), log))
    l.add_listener(Recorder(lambda e: "b:" + e.kind, log))
    t = l.park(car("C1"))
    try:
        l.park(car("C1"))
    except ParkingStateError:
        pass
    l.unpark(t.id)
    eq(["a:PARKED:F1-M1", "b:PARKED", "a:LEFT:F1-M1", "b:LEFT"], log, "events; failed park emits nothing")


def s5_display_board_tracks_availability():
    l = lot()
    pre = l.park(bike("B0"))
    board = DisplayBoard(l)
    eq(2, board.free(SpotSize.SMALL), "board starts from current availability")
    c = l.park(car("C1"))
    l.park(truck("TR1"))
    l.unpark(pre.id)
    l.unpark(c.id)
    for s in SpotSize:
        eq(l.available(s), board.free(s), "board matches lot for " + s.name)


# ---------------- step 6 ----------------
def s6_concurrent_parking_never_double_assigns():
    l = ParkingLot(HourlyPricing(RATES), NearestSpotSelector(), FakeClock())
    l.add_floor(1, 0, 10, 0)
    l.add_floor(2, 0, 10, 0)
    board = DisplayBoard(l)
    start = threading.Barrier(32)
    spots = []
    spots_lock = threading.Lock()

    def gate(i):
        start.wait()
        try:
            sid = l.park(car("C%d" % i)).spot_id
            with spots_lock:
                spots.append(sid)
        except ParkingFullError:
            pass

    threads = [threading.Thread(target=gate, args=(i,)) for i in range(32)]
    for t in threads:
        t.start()
    for t in threads:
        t.join(10)
    eq(20, len(spots), "exactly 20 parks succeed")
    eq(20, len(set(spots)), "20 distinct spots")
    eq(0, l.available(SpotSize.MEDIUM), "lot full")
    eq(0, board.free(SpotSize.MEDIUM), "board consistent under concurrency")
