from solution import BookingService, PaymentFailedError, PaymentGateway


class CapGateway(PaymentGateway):
    def charge(self, user, amount):
        return amount < 2_000


now = [0]
bms = BookingService(CapGateway(), lambda: now[0], 5 * 60_000)
bms.add_theatre("pvr-koramangala", "Bengaluru")
bms.add_screen("pvr-koramangala", "audi-1", 2, 1, 4)
bms.add_movie("m1", "Interstellar", 169)
show = bms.add_show("m1", "audi-1", 3_600_000, 250, 400)
print("shows in Bengaluru:", bms.find_shows("bengaluru", "m1"))
hold = bms.hold(show, "asha", ["C1", "C2", "A1"])
print("held", hold, ", available now", len(bms.available_seats(show)))
b = bms.confirm(hold)
print(b)
try:
    bms.confirm(bms.hold(show, "ravi", ["A2", "A3", "A4", "B1", "B2", "B3", "B4", "C3"]))
except PaymentFailedError:
    print("payment failed -> seats released, available", len(bms.available_seats(show)))
print(bms.cancel(b.id))
