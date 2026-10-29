Design an **immutable** `HttpRequest` that is always valid once constructed. It's a classic warm-up, and in Go it's the place to show you know *functional options*.

## Rules

| Field | Default | Rule (violations are errors) |
|---|---|---|
| url | (required) | must start with `http://` or `https://` |
| method | `GET` | one of `GET POST PUT PATCH DELETE`, case-insensitive input, stored upper-case |
| headers | empty | names are **case-insensitive**: stored lower-case, and a later value replaces an earlier one |
| timeout | 30 000 ms | 1 … 60 000 ms |
| body | none (`null` / `""`) | **not allowed** for `GET` and `DELETE` |

After construction, nothing can change the request. Mutating the builder afterwards, or the map you get from `headers()`, must not affect it.

<div class="lang-note" data-lang="java">

**Java API:** `HttpRequest.builder(url)` returns a `Builder` with chainable `method(String)`, `header(String, String)`, `timeoutMillis(long)`, `body(String)`, and `build()`.
Getters: `method()`, `url()`, `headers()` (unmodifiable `Map<String,String>`), `timeoutMillis()`, `body()`.
Errors: invalid method or timeout throws `IllegalArgumentException` **from the setter**. A bad URL or a body on GET/DELETE throws `IllegalStateException` **from `build()`**.

</div>

<div class="lang-note" data-lang="go">

**Go API:** `NewRequest(url string, opts ...Option) (*Request, error)` with `WithMethod(string)`, `WithHeader(k, v string)`, `WithTimeout(time.Duration)`, `WithBody(string)`.
Getters: `Method()`, `URL()`, `Header(name) (string, bool)`, `Headers() map[string]string` (a **copy**), `Timeout()`, `Body()`.
Errors (use `errors.Is`): `ErrInvalidURL`, `ErrInvalidMethod`, `ErrInvalidTimeout`, `ErrBodyNotAllowed`.

</div>

<div class="lang-note" data-lang="python">

**Python API:** `HttpRequest.builder(url)` returns a `Builder` with chainable `method(str)`, `header(name, value)`, `timeout_ms(int)`, `body(str)`, and `build()`.
The request exposes read-only properties: `method`, `url`, `headers` (a read-only mapping, so writing to it raises `TypeError`), `timeout_ms`, `body`.
Errors: an invalid method or timeout raises `ValueError` **from the setter**. A bad URL or a body on GET/DELETE raises `ValueError` **from `build()`**.

In everyday Python you'd often use keyword arguments with defaults instead of a builder (`HttpRequest(url, method="POST", ...)`). Mention that trade-off if asked, but the builder is what this exercise tests.

</div>

## What the interviewer is looking for

- Defaults set in exactly one place.
- Validation in a predictable place: per field in setters/options, cross-field in `build`.
- Defensive copies. Immutability is a property you **prove**, not just claim.
