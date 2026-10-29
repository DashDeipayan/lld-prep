from solution import Location, NearestDriverStrategy, Rate, RideService, StandardPricing, VehicleClass

now = [0]
rates = {VehicleClass.MINI: Rate(50, 12, 2), VehicleClass.SEDAN: Rate(80, 15, 3), VehicleClass.SUV: Rate(120, 20, 4)}
uber = RideService(NearestDriverStrategy(10), StandardPricing(rates, 150), lambda: now[0])
uber.add_driver("d1", VehicleClass.MINI, Location(0, 0), 4.6)
uber.add_driver("d2", VehicleClass.MINI, Location(2, 1), 4.9)
uber.set_online("d1", True)
uber.set_online("d2", True)
trip = uber.request_ride("asha", Location(3, 1), Location(10, 5), VehicleClass.MINI)
print(trip, "assigned to", uber.driver_of(trip))
uber.arrived(trip)
uber.start(trip)
now[0] += 17 * 60_000 + 5_000
print("fare (1.5x surge) =", uber.end(trip), ", status", uber.status(trip).value)
