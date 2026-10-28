Build the core of a price-alert feature: a `StockTicker` subject, and a `ThresholdAlert` observer.

## StockTicker

- `subscribe(listener)` returns an **unsubscribe handle** (Java `Runnable`, Go `func()`). Calling it twice is harmless.
- `update(symbol, price)` notifies every listener with `(symbol, price)`, but **only if the price changed** for that symbol. The first price for a symbol always counts as a change.
- Listeners are notified **in subscription order**.
- A listener that throws (Java) or panics (Go) must not stop the others.
- A listener may unsubscribe **itself or others while being notified**. Listeners removed during a notification may still receive that event, but must not receive later ones.
- `listenerCount()` returns the number of active subscriptions.

## ThresholdAlert

`ThresholdAlert(symbol, threshold, above, sink)` appends a message to `sink` when the condition **becomes true**:

- `above = true`: condition is `price >= threshold`, message `"AAPL>=150 @151"`
- `above = false`: condition is `price <= threshold`, message `"AAPL<=100 @99"`

It's **edge-triggered**: once fired, it doesn't fire again until the condition has been false at least once (it re-arms). Prices for other symbols are ignored.

**Example** (`AAPL >= 150`): updates 149, 151, 155, 148, 150 → fires at **151** and **150** only.

<div class="lang-note" data-lang="go">

Go API: `NewStockTicker() *StockTicker`, `Subscribe(l PriceListener) func()`, `Update(symbol string, price int64)`, `ListenerCount() int`, `NewThresholdAlert(symbol string, threshold int64, above bool, sink *[]string) PriceListener`.

</div>
