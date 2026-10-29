from solution import (DisplayBoard, HourlyPricing, NearestSpotSelector, ParkingListener, ParkingLot, SpotSize, Vehicle,
                      VehicleType)


class PrintEvents(ParkingListener):
    def on_event(self, event):
        print("  event:", event)


now = [0]
lot = ParkingLot(HourlyPricing({VehicleType.BIKE: 10, VehicleType.CAR: 20, VehicleType.TRUCK: 50}),
                 NearestSpotSelector(), lambda: now[0])
lot.add_floor(1, 1, 2, 1)
lot.add_floor(2, 2, 2, 1)
board = DisplayBoard(lot)
lot.add_listener(PrintEvents())

bike = lot.park(Vehicle("KA-01-B", VehicleType.BIKE))
car = lot.park(Vehicle("KA-02-C", VehicleType.CAR))
truck = lot.park(Vehicle("KA-03-T", VehicleType.TRUCK))
print(bike, car, truck, sep="\n")
print("free medium (board) =", board.free(SpotSize.MEDIUM))

now[0] += 90 * 60_000  # 1h30
print(lot.unpark(car.id))
print("free medium (board) =", board.free(SpotSize.MEDIUM))
