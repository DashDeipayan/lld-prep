**Interviewer:** *"Design an online auction. Sellers list items with an end time, buyers bid, the highest bid wins. Bidders get notified when they're outbid. Bonus: stop last-second sniping."*

## API

| Operation | Behaviour |
|---|---|
| `create(seller, item, startingPrice, minIncrement, endMillis)` → `"A1"`… | `endMillis` must be in the future |
| `subscribe(auctionId, listener)` | the listener receives every event for that auction, in order |
| `placeBid(auctionId, bidder, amount)` | see the rules below |
| `leader(id)`, `highest(id)` (0 if no bids), `endTime(id)` | queries |
| `close(id)` → `Result(id, winner, price)` | only once `now ≥ end`. Unsold → `winner = null`/`""`, `price = 0`. Calling it again returns the **same** result and emits nothing |

## Bid rules (in this order)

1. The auction must be open: `now < end` and not closed. Otherwise *auction closed*.
2. The seller can't bid → *invalid*.
3. The current leader can't bid again → *already leading*.
4. `amount ≥ startingPrice` for the first bid, otherwise `amount ≥ highest + minIncrement` → *bid too low*.

After an accepted bid:

- **Anti-sniping:** if fewer than **60 000 ms** remain, the end becomes `now + 60 000`.
- Events: `"BID:<bidder>:<amount>"`, then `"OUTBID:<previousLeader>"` if there was one.

On close: `"CLOSED:<winner>:<price>"` or `"CLOSED:UNSOLD"`.

## Concurrency (tested)

40 bidders each place one bid (amounts 100, 110, …, 490) at the same moment. Accepted bids must appear in the event stream in **strictly increasing** order, and the final leader is the one who bid 490.

<div class="lang-note" data-lang="java">

Provided: `interface Clock { long nowMillis(); }`, `interface AuctionListener { void onEvent(String auctionId, String event); }`, `record Result(String auctionId, String winner, long price)`, `BidRejectedException(String reason)`, where `reason` is one of `"CLOSED"`, `"LEADING"` or `"TOO_LOW"`. Seller bids and unknown ids throw `IllegalArgumentException`. Early close throws `IllegalStateException`. `leader` returns `Optional<String>`.

</div>

<div class="lang-note" data-lang="go">

Provided: `AuctionListener` interface, `ListenerFunc`, `Result{AuctionID, Winner string; Price int64}`. `NewAuctionHouse(clock func() time.Time)`. Times are `time.Time`. Errors: `ErrInvalid`, `ErrNotFound`, `ErrClosed`, `ErrAlreadyLeading`, `ErrBidTooLow`, `ErrNotEnded`.

</div>
