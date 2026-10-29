from solution import Hotel, RoomType

h = Hotel()
h.add_room("101", RoomType.SINGLE, 3_000)
h.add_room("102", RoomType.DOUBLE, 4_500)
h.add_room("201", RoomType.SUITE, 9_000)
print("availability:", {t.value: n for t, n in h.availability(10, 12).items()})
b = h.book("asha", RoomType.DOUBLE, 10, 12)
print(b, "-> room", h.room_of(b))
h.check_in(b, 10)
h.add_charge(b, 750)
print("bill =", h.check_out(b), "(%s)" % h.status(b))
