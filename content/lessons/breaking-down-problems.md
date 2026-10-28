---
title: How to Break Down Any LLD Problem
summary: A repeatable method from a vague prompt to classes, APIs and code, plus the project layout that machine-coding interviewers expect.
---

From today on, every day is about solving whole problems. This lesson is the method you'll apply to each of them. Read it once now and come back before every mock.

## The worked method, on "Design a Library Management System"

### 1. Turn the prompt into numbered use cases

The prompt is deliberately vague. Write down what you'll build and confirm it with the interviewer:

1. A librarian adds books. A book can have several physical **copies**.
2. A member **borrows** a copy, up to 5 at a time, for 14 days.
3. A member **returns** a copy, and a late return incurs a **fine** (₹10/day).
4. A member can **reserve** a book with no free copy, and gets notified when one is returned.
5. Search by title, author or ISBN.

**Out of scope (say it):** payments, authentication, UI, persistence. Repositories are in-memory.

### 2. Nouns become candidates; classify them

| Noun | Kind | Notes |
|---|---|---|
| Book | entity | ISBN, title, authors. It's the **catalogue entry** |
| BookCopy | entity | barcode, status. This is what's actually lent |
| Member | entity | id, borrowing limit |
| Loan | entity | copy, member, dueDate, returnedAt. Its **lifecycle** is ACTIVE → RETURNED |
| Reservation | entity | member, book, createdAt. A FIFO queue per book |
| Fine | value / derived | computed from a loan, not stored state |
| Librarian | **actor**, not a class | they call the service |
| Library | service (facade) | `LibraryService` |

The Book vs BookCopy split is the insight interviewers look for. **One noun in the prompt is often two classes**: a catalogue item and an instance (Movie vs Show, Product vs InventoryItem, Room type vs Room).

### 3. Find the points of variation and make them interfaces

Ask *"what will product managers change next quarter?"*

- the fine policy (per-day, capped, free for faculty) → `FinePolicy` **Strategy**
- the notification channel → `Notifier` interface (**Observer** on returns)
- search → `SearchIndex` (simple in-memory maps for now)

Everything else stays concrete. **Don't abstract what doesn't vary.** It's scored as over-engineering.

### 4. Write the service API first

```java
class LibraryService {
    Loan borrow(String memberId, String isbn);           // picks any available copy
    Money giveBack(String barcode);                      // returns the fine, if any
    Reservation reserve(String memberId, String isbn);
    List<Book> search(String query);
}
```

```go
type LibraryService interface {
	Borrow(memberID, isbn string) (*Loan, error)
	Return(barcode string) (fine Money, err error)
	Reserve(memberID, isbn string) (*Reservation, error)
	Search(query string) []Book
}
```

### 5. Code top-down, then the details

```java
Loan borrow(String memberId, String isbn) {
    Member m = members.get(memberId);                     // throws NotFound
    if (loans.activeCount(m) >= m.limit()) throw new LimitReachedException();
    BookCopy copy = copies.firstAvailable(isbn)           // throws NoCopyAvailable
            .orElseThrow(() -> new NoCopyAvailableException(isbn));
    copy.markBorrowed();
    Loan loan = new Loan(ids.next(), copy, m, clock.today().plusDays(14));
    loans.save(loan);
    return loan;
}
```

Every line calls something with an obvious name. You can now fill in `firstAvailable` and `activeCount` in any order, or skip them if time runs out, and the design still reads as complete.

### 6. Edge cases and concurrency pass

- Borrowing the last copy at the same time as someone else → lock per book, or a CAS on the copy status
- Returning a copy that isn't on loan → error
- A reserved book returned → hand it to the first reserver (hold it for 2 days), don't put it back on the shelf
- Fine boundaries: returned on the due date costs 0

## Project layout for machine-coding rounds

Even in one file, organise your code like a small project. Interviewers scan the structure first.

```text
model/        Book, BookCopy, Member, Loan, Reservation, enums
repository/   BookRepository (Map-backed), LoanRepository ...
strategy/     FinePolicy, PerDayFine
service/      LibraryService (facade), NotificationService
exception/    NotFoundException, LimitReachedException ...
Main          builds everything (dependency injection by hand) and runs the demo
```

```go
// One package is fine in Go for an interview. Group by file instead:
// model.go, repository.go, policy.go, service.go, errors.go, main.go
```

**Dependency injection by hand:** `main` constructs the repositories, strategies and clock, and passes them into the service constructors. No frameworks, no singletons.

## Time management cheat-sheet

| Elapsed (45 min OOD) | You should be… |
|---|---|
| 5 min | done with requirements, scope stated |
| 10 min | done with entities, a sketch of the class diagram and the service API |
| 30 min | done with the main flow coded end to end |
| 38 min | done with edge cases and concurrency |
| 45 min | done discussing extensions |

For a **90-minute machine-coding** round: requirements by 10 minutes, a working core by 55, a demo-ready `main` by 75, then tests and extensions.

## Red flags to avoid

- Starting with the database schema or a REST API (that's HLD, not LLD)
- `Manager` / `Helper` / `Util` classes with no clear responsibility
- Inheritance hierarchies more than two levels deep
- Every class a Singleton
- Getters and setters on everything with no behaviour in the entities
- Silent failure: returning `null` or `false` without saying why
