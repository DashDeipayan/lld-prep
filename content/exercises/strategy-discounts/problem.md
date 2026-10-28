A checkout service started with a single `if/else` chain for promotions. Every new promotion means editing that method, and it now has merge conflicts every sprint. Redesign it so **adding a promotion never touches `Checkout`**.

All money is in **cents** (integers) to avoid floating-point rounding.

## Requirements

`Cart` is already implemented for you. Implement:

| Strategy | Behaviour |
|---|---|
| `PercentageOff(percent)` | `subtotal * percent / 100` (integer division) |
| `FlatOff(amount, minSubtotal)` | `amount` off, but only when `subtotal >= minSubtotal`, otherwise 0 |
| `BuyNGetOneFree(sku, n)` | For that SKU, every group of `n + 1` units has one free unit (the unit price of that SKU) |

`Checkout` is built with any number of strategies. `total(cart)` returns `subtotal - bestDiscount`, where **bestDiscount is the single largest discount** (promotions don't stack). The total never goes below `0`. With no strategies, the total is the subtotal.

## Example

Cart: 3 × `BOOK` at 500, 1 × `PEN` at 100 → subtotal **1600**

- `PercentageOff(10)` → 160
- `FlatOff(300, 1500)` → 300
- `BuyNGetOneFree("BOOK", 2)` → 500

`Checkout([all three]).total(cart)` = **1100**

## What the interviewer is looking for

- One interface, one responsibility per strategy. That's Open/Closed in practice.
- `Checkout` depends on the abstraction, not on concrete classes (Dependency Inversion).
- Edge cases: empty cart, a SKU that isn't in the cart, a discount bigger than the subtotal.

<div class="lang-note" data-lang="go">

In Go, use constructor functions (`NewPercentageOff`, …) that return the `DiscountStrategy` interface, and `NewCheckout(strategies ...DiscountStrategy)`.

</div>
