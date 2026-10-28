Checkout supports card, UPI and wallet, and finance wants to add BNPL next quarter. Design a factory that creates a processor by method name, where **adding a method means registering it, not editing a `switch`**.

## Requirements

**Processors** implement `PaymentProcessor`. `pay(amountCents)` returns a `Receipt(method, amount, fee)` and rejects `amount <= 0`.

| Processor | method name | fee |
|---|---|---|
| Card | `card` | 2% (`amount * 2 / 100`) |
| UPI | `upi` | 0 |
| Wallet | `wallet` | 1% (`amount / 100`) |

**Factory**

- `register(method, constructor)`: method names are **case-insensitive**. Registering the same name twice is an error.
- `create(method)` returns a **new** processor instance each call. An unknown method is an error.
- `supportedMethods()` returns the registered names, lower-case and **sorted**.
- `PaymentFactory.withDefaults()` / `NewDefaultFactory()` returns a factory with the three processors above already registered.

<div class="lang-note" data-lang="java">

Errors: duplicate register throws `IllegalStateException`. Unknown method or `amount <= 0` throws `IllegalArgumentException`.

</div>

<div class="lang-note" data-lang="go">

`NewPaymentFactory() *PaymentFactory`, `Register(method string, ctor func() PaymentProcessor) error`, `Create(method string) (PaymentProcessor, error)`, `SupportedMethods() []string`, `NewDefaultFactory() *PaymentFactory`. Errors: `ErrDuplicateMethod`, `ErrUnsupportedMethod`, `ErrInvalidAmount`.

</div>

## Talking points

- Why a registry instead of a `switch`? (Open/Closed. Plugins can register themselves.)
- Why store constructors instead of instances? (Processors may hold per-transaction state.)
- How would you make `register` safe to call from several threads at startup?
