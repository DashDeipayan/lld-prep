Build the scheduling core of an elevator bank. Read the elevator walkthrough lesson first. Floors are `0 … floors−1`. All cars start at floor 0, `IDLE`.

## Time is discrete: `step()`

Every `step()` processes the cars **in id order**. For each car:

1. No stops → direction becomes `IDLE`, nothing happens.
2. If the current floor is a stop → remove it and **open doors** (record a stop event). **Don't move** this tick.
3. If `IDLE` → head toward the **nearest** stop (`UP` on ties).
4. If heading `UP` with no stop above → switch to `DOWN`. If heading `DOWN` with no stop below → switch to `UP`. (This is LOOK.)
5. Move one floor. If the new floor is a stop → remove it and record a stop event.

Whenever a car removes its **last** stop, its direction becomes `IDLE`. `step()` returns this tick's stop events as `"E<id>@<floor>"`, in id order.

## Requests

- `requestDestination(carId, floor)`: a button pressed *inside* a car adds a stop to that car.
- `requestPickup(floor, dir)`: a hall button. The **dispatch strategy** chooses a car, which gets `floor` as a stop. Returns the car id.

## Default strategy: `NearestCarStrategy`

```text
cost(car) =
  car IDLE                                          → |car.floor − floor|
  car moving UP, dir UP, floor ≥ car.floor          → floor − car.floor
  car moving DOWN, dir DOWN, floor ≤ car.floor      → car.floor − floor
  otherwise → |car.floor − turn| + |turn − floor|,  where turn = car's highest stop if UP, lowest if DOWN
lowest cost wins; ties → lowest id
```

## Example (1 car)

Destinations 5, 2, 8 from floor 0 → stop events at ticks 2 (`E0@2`), 5 (`E0@5`), 8 (`E0@8`).

<div class="lang-note" data-lang="java">

Java: `Direction { UP, DOWN, IDLE }` and `DispatchStrategy { Elevator choose(List<Elevator> cars, int floor, Direction dir); }` are provided. `Elevator` exposes `id()`, `floor()`, `direction()` and `stops()` (sorted `List<Integer>`). `ElevatorSystem(floors, cars, strategy)`. An invalid floor, car id or pickup direction (`IDLE`) throws `IllegalArgumentException`.

</div>

<div class="lang-note" data-lang="go">

Go: `Direction` (`Up`, `Down`, `Idle`) and `DispatchStrategy { Choose(cars []*Elevator, floor int, dir Direction) *Elevator }` are provided. `*Elevator` exposes `ID()`, `Floor()`, `Dir()` and `Stops() []int` (sorted). `NewElevatorSystem(floors, cars int, s DispatchStrategy) *ElevatorSystem`, `RequestPickup(floor, dir) (int, error)`, `RequestDestination(id, floor) error`, `Step() []string`, `Car(id) *Elevator`. Errors: `ErrInvalidFloor`, `ErrInvalidElevator`, `ErrInvalidDirection`.

</div>

<div class="lang-note" data-lang="python">

Python: the `Direction` enum and the `DispatchStrategy` abstract base class (`choose(cars, floor, direction)`) are provided. `Elevator` exposes the properties `id`, `floor` and `direction`, plus `stops()` (a sorted list). Use `ElevatorSystem(floors, cars, strategy)` with `request_pickup(floor, direction)`, `request_destination(car_id, floor)`, `step()` and `car(car_id)`. An invalid floor, car id or `IDLE` pickup raises `ValueError`.

</div>
