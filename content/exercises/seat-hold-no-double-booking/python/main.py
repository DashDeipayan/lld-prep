from solution import SeatInventory, SeatUnavailableError

MINUTE = 60_000
now = [0]
inv = SeatInventory(["A1", "A2", "A3", "A4"], 5 * MINUTE, lambda: now[0])
h1 = inv.hold("alice", ["A1", "A2"])
print("alice holds", h1, ", available", inv.available())
try:
    inv.hold("bob", ["A2", "A3"])
except SeatUnavailableError as e:
    print("bob failed on", e, "(A3 untouched: %s)" % inv.status("A3"))
now[0] += 6 * MINUTE
print("after expiry, available", inv.available())
h2 = inv.hold("bob", ["A2", "A3"])
inv.confirm(h2)
print("bob booked: A2=%s, A3=%s" % (inv.status("A2"), inv.status("A3")))
