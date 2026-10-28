---
title: OOP & SOLID, the Interview Edition
summary: You know the definitions. This is how interviewers probe them, the smells they look for, and the one-line justifications that score.
---

You've used classes and interfaces for years, so this lesson skips the definitions. It covers the **decisions** an LLD interviewer watches you make, and how to **say them out loud**.

## The four decisions that come up in every round

### 1. Composition vs inheritance

Default to **composition**. Use inheritance only for a true *is-a* relationship where the subclass honours the parent's contract everywhere (see LSP below).

The classic trap: `class ElectricCar extends Car`, `class SelfDrivingElectricCar extends ElectricCar`… Every new combination of features adds another class. Model the *varying behaviour* as a component instead:

```java
class Car {
    private final Engine engine;          // has-a: Petrol, Electric, Hybrid
    private final DrivingMode mode;       // has-a: Manual, Autopilot
    Car(Engine engine, DrivingMode mode) { this.engine = engine; this.mode = mode; }
    void drive() { mode.steer(); engine.propel(); }
}
```

```go
// Go has no inheritance at all: composition is the only option.
type Car struct {
    engine Engine      // interface
    mode   DrivingMode // interface
}

func (c *Car) Drive() { c.mode.Steer(); c.engine.Propel() }
```

**Say:** *"I'm composing the engine instead of subclassing Car, so engine type and driving mode vary independently. With inheritance that would be N×M classes."*

<div class="lang-note" data-lang="go">

**Go note:** struct embedding (`type Admin struct{ User }`) promotes methods, but it's **not** subtyping. You can't pass an `Admin` where a `User` is expected. Polymorphism in Go comes from interfaces only, and they're satisfied implicitly. Interviewers who know Go like to hear you say that.

</div>

### 2. Interface vs abstract class

- **Interface:** a capability or contract that several unrelated types can fulfil (`PaymentMethod`, `Notifier`). This is the default choice.
- **Abstract class (Java):** when subclasses share **state and a fixed algorithm skeleton** (Template Method). For example, `abstract class Vehicle { String plate; abstract SpotSize size(); }`.

In Go, use an interface for the contract and a small embedded struct for shared fields or helpers.

### 3. Encapsulation: tell, don't ask

Interviewers flag code that pulls data out of an object to make a decision *about* that object.

```java
// ✗ Asking: logic about Account lives outside Account
if (account.getBalance() >= amount) account.setBalance(account.getBalance() - amount);

// ✓ Telling: Account protects its own invariant
account.withdraw(amount);   // throws InsufficientFundsException
```

```go
// ✓ The invariant lives with the data
func (a *Account) Withdraw(amount int64) error {
    if amount > a.balance {
        return ErrInsufficientFunds
    }
    a.balance -= amount
    return nil
}
```

Rule: **no public setters on entities** unless the field really is free-form. Each state change gets a method named after the domain operation (`checkIn()`, `cancel()`, `assignDriver()`).

### 4. Entities vs value objects

- **Entity:** has identity and a lifecycle: `Order#123` stays the same order as it changes status.
- **Value object:** immutable and compared by value: `Money(500, INR)`, `Location(lat, lng)`, `TimeSlot(start, end)`.

Value objects remove whole categories of bugs (a primitive `long price` can't say which currency it's in) and make method signatures self-documenting. Use Java `record`s and small Go structs passed by value.

## SOLID: one smell, one fix each

Interviewers rarely ask you to recite SOLID. They **show you code that violates it** or watch whether your own code does. Learn to spot each smell.

### S: Single Responsibility

**Smell:** a class whose description needs the word "and". `OrderService` validates, prices, charges cards, sends email *and* writes to the DB.

**Fix:** split by **reason to change**. Pricing rules change when marketing asks. Email changes when the template changes. Each becomes its own class that the coordinator calls.

### O: Open/Closed

**Smell:** a `switch`/`if-else` on a type code that you edit every time a new type arrives.

```java
double fee(Vehicle v) {
    switch (v.type()) {             // ✗ add TRUCK → edit this method
        case CAR: return 20;
        case BIKE: return 10;
    }
}
```

**Fix:** polymorphism or a strategy, so a new type is a **new class** and no existing code is edited. The pattern lessons are mostly this principle in different shapes.

### L: Liskov Substitution

**Smell:** a subclass that throws `UnsupportedOperationException`, or weakens a guarantee of the parent. The textbook case is `Square extends Rectangle` breaking `setWidth`. The interview case is `class Penguin extends Bird` with `fly()` throwing an exception.

**Fix:** narrow the abstraction. Split `Bird` and `FlyingBird`, or model the capability as an interface (`Flyable`) that only some types implement.

**Say:** *"If a caller holding the base type has to check which subtype it got, LSP is broken."*

### I: Interface Segregation

**Smell:** a fat interface where implementers stub half the methods. `interface Machine { print(); scan(); fax(); }` forces `BasicPrinter` to fake `fax()`.

**Fix:** small, role-based interfaces: `Printer`, `Scanner`. Go's standard library is the model to follow: `io.Reader`, `io.Writer`, one method each.

### D: Dependency Inversion

**Smell:** high-level policy creates its low-level details: `new SmtpEmailSender()` inside `OrderService`. You can't test it and you can't swap it.

**Fix:** depend on an interface and **inject** the implementation through the constructor.

```java
class OrderService {
    private final Notifier notifier;              // abstraction
    OrderService(Notifier notifier) { this.notifier = notifier; }
}
```

```go
type OrderService struct {
    notifier Notifier // interface, injected
}

func NewOrderService(n Notifier) *OrderService { return &OrderService{notifier: n} }
```

Constructor injection is also how you'll make every exercise in this course **testable**: tests inject fakes (a fixed clock, a deterministic random source, a recording notifier).

## Smells interviewers comment on, beyond SOLID

| Smell | Example | Better |
|---|---|---|
| Primitive obsession | `String status`, `int type` | `enum OrderStatus`, `enum VehicleType` |
| Stringly-typed IDs everywhere | `book(String, String, String)` | small types or clearly named params |
| God manager | `SystemManager` with 40 methods | one service per aggregate |
| Anemic model | entities with only getters/setters, all logic in services | put invariants on the entity |
| Singleton for everything | `ParkingLot.getInstance()` used inside entities | create once in `main`, inject |
| Hidden time & randomness | `System.currentTimeMillis()`/`time.Now()` deep inside logic | inject a `Clock`/`func() time.Time` |
| Exceptions for control flow / swallowed errors | `catch (Exception e) {}` | domain exceptions / sentinel errors |

## Errors: Java vs Go conventions

- **Java:** throw **domain-specific unchecked exceptions** (`SpotNotAvailableException extends RuntimeException`). Return `Optional<T>` for "may be absent" lookups.
- **Go:** return `error` as the last value. Declare **sentinel errors** (`var ErrLotFull = errors.New("lot full")`) so callers can use `errors.Is`. Wrap with context using `fmt.Errorf("park %s: %w", plate, ErrLotFull)`.

Either way, pick one convention and use it consistently. Inconsistent error handling is one of the easiest code-quality points to lose.
