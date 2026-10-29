import threading

from lldtest import eq, raises
from solution import ConflictError, NoRoomAvailableError, Scheduler, Slot


def sched():
    s = Scheduler()
    s.add_room("R1", 4)
    s.add_room("R2", 8)
    s.add_room("R0", 4)
    s.add_room("R3", 2)
    return s


def test_best_fit_room_selection():
    s = sched()
    eq("R0", s.book("a", 600, 660, 3).room, "smallest capacity >= 3, tie -> lowest id (R0)")
    eq("R1", s.book("b", 600, 660, 3).room, "R0 busy -> R1")
    eq("R2", s.book("c", 600, 660, 3).room, "both 4-seaters busy -> R2")
    raises(NoRoomAvailableError, lambda: s.book("d", 600, 660, 3), "nothing left that fits")
    eq("R3", s.book("e", 600, 660, 2).room, "2-seater still free")


def test_half_open_intervals():
    s = sched()
    s.book_room("R1", "a", 600, 660)
    eq("M2", s.book_room("R1", "b", 660, 720).id, "back-to-back is fine")
    raises(ConflictError, lambda: s.book_room("R1", "c", 659, 661), "overlap")
    raises(ConflictError, lambda: s.book_room("R1", "c", 500, 800), "enclosing overlap")
    raises(ConflictError, lambda: s.book_room("R1", "c", 610, 620), "inside overlap")
    eq("M3", s.book_room("R1", "c", 540, 600).id, "ends exactly at next start")


def test_validation():
    s = sched()
    raises(ValueError, lambda: s.book("a", 600, 600, 2), "empty interval")
    raises(ValueError, lambda: s.book("a", -1, 60, 2), "negative")
    raises(ValueError, lambda: s.book("a", 1400, 1441, 2), "past midnight")
    raises(ValueError, lambda: s.book("a", 0, 60, 0), "no attendees")
    raises(ValueError, lambda: s.book_room("RX", "a", 0, 60), "unknown room")
    raises(ValueError, lambda: s.add_room("R1", 3), "duplicate room")


def test_cancel_and_schedule():
    s = sched()
    b1 = s.book_room("R2", "a", 700, 760)
    b2 = s.book_room("R2", "b", 600, 630)
    eq([b2, b1], s.schedule("R2"), "sorted by start")
    s.cancel(b1.id)
    eq([b2], s.schedule("R2"), "cancelled removed")
    raises(ConflictError, lambda: s.cancel(b1.id), "double cancel")
    eq("R2", s.book_room("R2", "c", 700, 760).room, "slot reusable after cancel")


def test_free_rooms_and_slots():
    s = sched()
    s.book_room("R0", "a", 600, 660)
    eq(["R3", "R1", "R2"], s.free_rooms(630, 690, 2), "fit + free, ordered by capacity then id")
    s.book_room("R1", "a", 540, 570)
    s.book_room("R1", "b", 600, 630)
    s.book_room("R1", "c", 630, 700)
    eq([Slot(480, 540), Slot(570, 600), Slot(700, 720)], s.free_slots("R1", 480, 720), "gaps, adjacent bookings merge")
    eq([], s.free_slots("R1", 600, 700), "fully booked window")
    eq([Slot(0, 1440)], s.free_slots("R2", 0, 1440), "empty room")


def test_single_winner_under_contention():
    s = sched()
    start = threading.Barrier(50)
    wins = []

    def attempt(who):
        start.wait()
        try:
            s.book_room("R1", who, 600, 660)
            wins.append(who)
        except ConflictError:
            pass

    threads = [threading.Thread(target=attempt, args=("u%d" % i,)) for i in range(50)]
    for t in threads:
        t.start()
    for t in threads:
        t.join(10)
    eq(1, len(wins), "exactly one booking wins")
    eq(1, len(s.schedule("R1")), "one booking stored")
