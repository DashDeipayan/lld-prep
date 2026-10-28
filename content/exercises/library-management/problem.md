The problem from the *How to Break Down Any LLD Problem* lesson, now fully specified and tested. Time is in **day numbers** (an integer day counter) from an injected clock.

## Catalogue & members

- `addBook(isbn, title, author)`, `addCopy(isbn, barcode)`, `addMember(memberId, maxLoans)`. Duplicates and unknown references are errors.
- `availableCopies(isbn)` counts copies anyone can borrow right now.
- `search(query)` returns ISBNs whose **title or author** contains the query, case-insensitively, sorted by **title**.

## Borrowing

`borrow(memberId, isbn)` → `Loan(id, barcode, isbn, memberId, dueDay)`

- The member may not exceed `maxLoans` active loans.
- Copy choice: a copy **held for this member** first, otherwise the **lowest barcode** among available copies.
- `dueDay = today + 14`. Loan ids are `"L1"`, `"L2"`, …
- No copy available to this member → error.

## Returning

`giveBack(barcode)` → fine, where `daysLate = max(0, today − dueDay)` and the fine comes from the injected **FinePolicy** strategy. `PerDayFine(perDay, cap)` returns `min(cap, daysLate × perDay)`.

Then: if someone has **reserved** this ISBN, the copy is **held** for the first member in the FIFO queue until `today + 2` (inclusive), and the **Notifier** is called with `(memberId, "<isbn> is ready for pickup")`. Otherwise the copy becomes available. Returning a copy that isn't on loan is an error.

## Reservations

`reserve(memberId, isbn)` is only allowed when **no copy is available** (otherwise just borrow it). The same member can't reserve the same ISBN twice. A hold that isn't picked up by the end of day `today + 2` lapses: the copy becomes available to anyone, and the reservation is consumed.

<div class="lang-note" data-lang="java">

Java: `interface Clock { long today(); }`, `interface FinePolicy { long fine(long daysLate); }`, `interface Notifier { void notify(String memberId, String message); }` and `record Loan(...)` are provided. Constructor `Library(FinePolicy, Clock, Notifier)`. Unknown or duplicate entities throw `IllegalArgumentException`. Rule violations (limit, nothing available, not on loan, reserving when available, duplicate reservation) throw `IllegalStateException`.

</div>

<div class="lang-note" data-lang="go">

Go: `type FinePolicy interface { Fine(daysLate int) int64 }`, `type Notifier interface { Notify(memberID, msg string) }` and `type Loan struct { ID, Barcode, ISBN, MemberID string; DueDay int }` are provided, plus `NewPerDayFine(perDay, cap int64)`. `NewLibrary(f FinePolicy, today func() int, n Notifier)`. Errors: `ErrNotFound`, `ErrDuplicate`, `ErrLimitReached`, `ErrNoCopyAvailable`, `ErrNotOnLoan`, `ErrCopyAvailable`, `ErrAlreadyReserved`.

</div>
