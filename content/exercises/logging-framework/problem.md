"Design a logging framework" is one of the most frequently asked LLD questions. Build the core.

## Concepts

- **Level:** `DEBUG < INFO < WARN < ERROR`
- **Logger:** has a `name` and a minimum level that can be changed at runtime. `debug/info/warn/error(msg)` drops anything below the logger's level, then passes the record to **every** handler, in the order they were added.
- **Handler:** has its **own** minimum level and a **Formatter**. It writes only records at or above its level. A handler that throws must not affect the others.
- **Formatter** (a Strategy):
  - `SimpleFormatter` → `WARN [orders] payment slow`
  - `JsonFormatter` → `{"level":"WARN","logger":"orders","msg":"payment slow"}`. Escape `"` as `\"` and `\` as `\\`.

`MemoryHandler(minLevel, sink, formatter)` appends formatted lines to `sink`. It stands in for console and file handlers.

## Example

```text
logger "orders" at INFO
  ├─ MemoryHandler(DEBUG, console, Simple)
  └─ MemoryHandler(ERROR, alerts,  Json)

info("created #1")   → console: INFO [orders] created #1
debug("cache miss")  → dropped by the logger
error("charge failed") → console + alerts
```

## Requirements checked by tests

1. Level filtering at both the logger and the handler.
2. Handlers run in the order added. A throwing handler is isolated.
3. `setLevel` takes effect immediately.
4. 8 threads × 500 messages through one logger produces exactly 4 000 lines.

<div class="lang-note" data-lang="go">

Go API: `type Level int` with constants `Debug, Info, Warn, Error` and a `String()` method that returns `"DEBUG"` etc. `NewLogger(name string, min Level) *Logger`. `(*Logger).AddHandler(h Handler) *Logger`. `SetLevel(Level)`. `Debug/Info/Warn/Error(msg string)`. `Handler` interface: `Handle(logger string, lvl Level, msg string) error`. `Formatter` interface: `Format(logger string, lvl Level, msg string) string`. `NewMemoryHandler(min Level, sink *[]string, f Formatter) Handler`. Formatters: `SimpleFormatter{}`, `JSONFormatter{}`.

</div>

<div class="lang-note" data-lang="python">

Python API: `Level` is an `IntEnum`, so `Level.DEBUG < Level.INFO`, and `level.name` gives `"DEBUG"`. `Logger(name, min_level)` has `add_handler(h)` (which returns the logger, so calls chain), `set_level(level)` and `debug/info/warn/error(msg)`. `Handler` and `Formatter` are abstract base classes with `handle(logger, level, msg)` and `format(logger, level, msg)`. Use `MemoryHandler(min_level, sink, formatter)`. `json.dumps` can do the JSON escaping for you.

</div>
