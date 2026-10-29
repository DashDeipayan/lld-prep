import threading

from lldtest import eq, raises
from solution import BookingStatusError, Hotel, NoRoomAvailableError, RoomType

SINGLE, DOUBLE, SUITE = RoomType.SINGLE, RoomType.DOUBLE, RoomType.SUITE


def hotel():
    h = Hotel()
    h.add_room("102", DOUBLE, 4000)
    h.add_room("101", DOUBLE, 4000)
    h.add_room("103", SINGLE, 2500)
    h.add_room("301", SUITE, 9000)
    return h


def test_availability_and_assignment():
    h = hotel()
    eq({SINGLE: 1, DOUBLE: 2, SUITE: 1}, h.availability(1, 3), "all free")
    b1 = h.book("a", DOUBLE, 1, 3)
    eq("BK1", b1, "id")
    eq("101", h.room_of(b1), "lowest number first")
    eq("102", h.room_of(h.book("b", DOUBLE, 2, 4)), "overlapping stay gets the next room")
    eq({SINGLE: 1, DOUBLE: 0, SUITE: 1}, h.availability(2, 3), "doubles full on night 2")
    eq("101", h.room_of(h.book("c", DOUBLE, 3, 5)), "101 free from night 3")
    raises(NoRoomAvailableError, lambda: h.book("d", DOUBLE, 2, 4), "none left")
    raises(ValueError, lambda: h.book("d", SINGLE, 4, 4), "empty stay")


def test_lifecycle_and_bill():
    h = hotel()
    b = h.book("a", SUITE, 5, 8)
    raises(BookingStatusError, lambda: h.add_charge(b, 100), "no charges before check-in")
    raises(BookingStatusError, lambda: h.check_in(b, 4), "wrong day")
    h.check_in(b, 5)
    eq("CHECKED_IN", h.status(b), "checked in")
    h.add_charge(b, 1200)
    h.add_charge(b, 300)
    raises(ValueError, lambda: h.add_charge(b, 0), "positive charge")
    eq(3 * 9000 + 1500, h.check_out(b), "3 nights + charges")
    eq("CHECKED_OUT", h.status(b), "checked out")
    raises(BookingStatusError, lambda: h.check_out(b), "double checkout")
    eq("301", h.room_of(h.book("b", SUITE, 6, 7)), "room freed after checkout")


def test_cancel():
    h = hotel()
    b = h.book("a", SINGLE, 1, 2)
    h.cancel(b)
    eq("CANCELLED", h.status(b), "cancelled")
    eq(1, h.availability(1, 2)[SINGLE], "room free again")
    raises(BookingStatusError, lambda: h.cancel(b), "double cancel")
    c = h.book("b", SINGLE, 1, 2)
    h.check_in(c, 1)
    raises(BookingStatusError, lambda: h.cancel(c), "can't cancel after check-in")


def test_last_suite_single_winner():
    h = hotel()
    start = threading.Barrier(20)
    ok = []

    def attempt(i):
        start.wait()
        try:
            h.book("g", SUITE, 10, 12)
            ok.append(i)
        except NoRoomAvailableError:
            pass

    threads = [threading.Thread(target=attempt, args=(i,)) for i in range(20)]
    for t in threads:
        t.start()
    for t in threads:
        t.join(10)
    eq(1, len(ok), "one suite, one booking")
