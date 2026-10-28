Implement the core of Uber/Ola. Read the ride-sharing walkthrough first. Locations are grid points, and distance is **Manhattan**: `|x1−x2| + |y1−y2|`.

## Drivers

- `addDriver(id, vehicleClass, location, rating)`. A new driver is **offline**.
- `setOnline(id, online)`. Going offline during a trip is an error.
- `updateLocation(id, location)`

## Matching: `NearestDriverStrategy(maxDistance)`

Among the candidates (online, available, requested class), choose the **nearest** within `maxDistance` (inclusive). Ties go to the **higher rating**, then the **lower id**. No candidate within range means no match.

## Trips

`requestRide(riderId, from, to, class)` → `"R1"`, `"R2"`, …

- A rider with an active (not completed or cancelled) trip can't request another.
- The matched driver becomes **busy** immediately, and the trip is `ASSIGNED`. No driver → *no driver available*.

Lifecycle: `arrived(trip)` → `ARRIVED`, `start(trip)` → `IN_PROGRESS` (the start time is recorded), `end(trip)` → `COMPLETED`, returning the **fare**. `cancel(trip)` is allowed only from `ASSIGNED` or `ARRIVED`. Any other transition is illegal. On `COMPLETED` or `CANCELLED` the driver becomes available again. On `COMPLETED` the driver's location becomes the trip's `to`.

`driverOf(trip)` and `status(trip)` are queries.

## Pricing: `StandardPricing(rates, surgePercent)`

```text
minutes = ceil(durationMs / 60 000)         distance = manhattan(from, to)
fare = (base + perKm × distance + perMin × minutes) × surgePercent / 100      (integer division)
```

## Concurrency (tested)

10 available drivers, 50 riders requesting at the same instant: exactly **10** trips are created, with **10 distinct** drivers.

<div class="lang-note" data-lang="java">

Provided: `enum VehicleClass { MINI, SEDAN, SUV }`, `enum TripStatus { ASSIGNED, ARRIVED, IN_PROGRESS, COMPLETED, CANCELLED }`, `record Location(int x, int y)` with `distance(other)`, `record DriverInfo(String id, VehicleClass cls, Location location, double rating)`, `record Rate(long base, long perKm, long perMin)`, `interface MatchingStrategy { Optional<String> match(Location pickup, List<DriverInfo> candidates); }`, `interface PricingStrategy { long fare(VehicleClass cls, int distance, long durationMillis); }`, `interface Clock`, and `NoDriverAvailableException`. Unknown ids throw `IllegalArgumentException`. Illegal transitions, a busy rider, or going offline mid-trip throw `IllegalStateException`.

</div>

<div class="lang-note" data-lang="go">

Provided: `VehicleClass` (`Mini`, `Sedan`, `SUV`), `TripStatus` (`Assigned`, `Arrived`, `InProgress`, `Completed`, `Cancelled`), `Location{X, Y}` with `Distance`, `DriverInfo`, `Rate`, `MatchingStrategy { Match(pickup Location, cands []DriverInfo) (string, bool) }`, `PricingStrategy { Fare(cls VehicleClass, distance int, d time.Duration) int64 }`. `NewNearestDriver(maxDistance int)`, `NewStandardPricing(rates map[VehicleClass]Rate, surgePercent int64)`, `NewRideService(m, p, clock func() time.Time)`. Errors: `ErrNotFound`, `ErrNoDriver`, `ErrIllegalTransition`, `ErrRiderBusy`, `ErrDriverBusy`.

</div>
