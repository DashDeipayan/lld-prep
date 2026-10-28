**Interviewer:** *"Design a car rental system like Zoomcar or Hertz. Customers search for cars in a city for a date range, reserve one, pick it up and return it. Let's code the core."*

## What you need to support

| Operation | Rules |
|---|---|
| `addBranch(branchId, city)` | |
| `addCar(branchId, carId, type, dailyRate)` | type ∈ HATCHBACK, SEDAN, SUV |
| `search(city, type, startDay, endDay)` | ids of cars in that city (case-insensitive) and type that are **free for the whole range**, sorted by **daily rate**, then id |
| `reserve(customer, carId, startDay, endDay)` → reservation | covers days `[start, end)`. `cost = (end − start) × rate`. Overlap with an active reservation → *conflict*. Ids `"RS1"`, `"RS2"`, … Status `"RESERVED"` |
| `cancel(reservationId)` | only while `RESERVED`. Status `"CANCELLED"`, and the car is free again |
| `returnCar(reservationId, actualDay)` → final amount | only while `RESERVED`. `final = cost + max(0, actualDay − endDay) × rate × 3 / 2`. Status `"COMPLETED"`, and the car is free from `actualDay` |

Invalid ranges (`start < 0`, `start ≥ end`), unknown ids and duplicates are invalid input.

**Concurrency:** 30 customers reserving the **same car for overlapping dates** at the same moment results in exactly **one** reservation.

<div class="lang-note" data-lang="java">

Provided: `enum CarType`, `record Reservation(String id, String customer, String carId, int startDay, int endDay, long cost, String status)`. Invalid input throws `IllegalArgumentException`. A conflict or wrong status throws `IllegalStateException`. Methods that return a reservation return a **snapshot** reflecting the new status.

</div>

<div class="lang-note" data-lang="go">

Provided: `CarType` (`Hatchback`, `Sedan`, `SUV`), `Reservation{ID, Customer, CarID string; StartDay, EndDay int; Cost int64; Status string}`. `NewRentalService()`. Methods return `(Reservation, error)` / `(int64, error)`. Errors: `ErrInvalid`, `ErrNotFound`, `ErrConflict`, `ErrBadStatus`.

</div>
