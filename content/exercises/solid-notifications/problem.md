The starter file contains `LegacyNotifications`, a class every team edits. It decides the channel with `if/else`, constructs its own SMTP and SMS clients, and if SMS fails, push is never attempted.

Redesign it so that:

1. **Adding a channel** (WhatsApp, Slack…) means adding one class, with **no edits** to the service.
2. Every dependency is **injected**, so the service can be tested without real gateways.
3. **One failing channel never blocks the others.**

## Requirements

`User` and `Outbox` are provided. `Outbox` stands in for the real gateways and just records what was sent as `"channel:to:message"`.

Implement a `Notifier` abstraction with three implementations:

| Notifier | `channel()` | Can reach the user when | Records |
|---|---|---|---|
| Email | `"email"` | `email` is non-empty | `email:<email>:<msg>` |
| SMS | `"sms"` | `phone` is non-empty | `sms:<phone>:<msg>` |
| Push | `"push"` | `deviceToken` is non-empty | `push:<token>:<msg>` |

`NotificationService` is constructed with any list of notifiers. `notify(user, message)`:

- rejects a blank message (Java: `IllegalArgumentException`, Go: `ErrEmptyMessage`)
- skips notifiers whose channel the user has **muted**, or that **can't reach** the user
- tries every remaining notifier **in order**. A notifier that throws or returns an error is skipped, and the rest still run
- returns the **number of successful sends**

## Follow-up to think about

*"Marketing wants to rate-limit SMS to one per user per hour."* Where does that go? (Answer: a decorator around the SMS notifier. You'll meet that pattern on Day 4.)

<div class="lang-note" data-lang="go">

Go signatures: `NewEmailNotifier(o *Outbox) Notifier` (likewise `NewSmsNotifier`, `NewPushNotifier`), `NewNotificationService(ns ...Notifier) *NotificationService`, and `(*NotificationService).Notify(u User, msg string) (int, error)`. `Notifier.Send` returns an `error`.

</div>
