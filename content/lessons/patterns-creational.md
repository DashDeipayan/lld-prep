---
title: Creational Patterns That Show Up in LLD Rounds
summary: Factory, Builder and Singleton, plus when interviewers expect them, when they're a red flag, and how each looks in idiomatic Go and Python.
---

Creational patterns answer one question: **who decides which concrete object gets built, and how?** In interviews, three come up all the time. Prototype and Abstract Factory almost never get past a one-line mention.

## Factory: centralise "which implementation?"

**Use when** the caller knows *what kind* of thing it needs (a payment method code, a vehicle type, a file extension) but shouldn't know the concrete class.

There are three variants. Know which one you're using:

| Variant | Shape | Use in interviews |
|---|---|---|
| **Simple factory** | one method with a `switch` that returns an interface | Fine for a closed set. Say it's a simple factory |
| **Registry factory** | `Map<key, Supplier>` filled at startup | **Best answer** when new types keep arriving (Open/Closed) |
| **Factory method** (GoF) | subclass overrides `createX()` | Rare. Mention it if a framework base class is involved |
| **Abstract factory** | factory of related families (`UIFactory` → button + checkbox) | Mention only. Almost never coded |

```java
interface PaymentProcessor { Receipt pay(long amountCents); }

final class PaymentFactory {
    private final Map<String, Supplier<PaymentProcessor>> registry = new HashMap<>();

    void register(String method, Supplier<PaymentProcessor> ctor) {
        if (registry.putIfAbsent(method, ctor) != null)
            throw new IllegalStateException("duplicate method " + method);
    }

    PaymentProcessor create(String method) {
        Supplier<PaymentProcessor> c = registry.get(method);
        if (c == null) throw new IllegalArgumentException("unsupported method " + method);
        return c.get();
    }
}
// startup: factory.register("UPI", UpiProcessor::new);
```

```go
type PaymentProcessor interface {
	Pay(amountCents int64) (Receipt, error)
}

type PaymentFactory struct {
	registry map[string]func() PaymentProcessor
}

func (f *PaymentFactory) Register(method string, ctor func() PaymentProcessor) error {
	if _, dup := f.registry[method]; dup {
		return fmt.Errorf("duplicate method %q", method)
	}
	f.registry[method] = ctor
	return nil
}

func (f *PaymentFactory) Create(method string) (PaymentProcessor, error) {
	ctor, ok := f.registry[method]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupported, method)
	}
	return ctor(), nil
}
```

```python
class PaymentProcessor(ABC):
    @abstractmethod
    def pay(self, amount_cents: int) -> Receipt: ...


class PaymentFactory:
    def __init__(self) -> None:
        self._registry: Dict[str, Callable[[], PaymentProcessor]] = {}

    def register(self, method: str, ctor: Callable[[], PaymentProcessor]) -> None:
        if method in self._registry:
            raise DuplicateMethodError("duplicate method " + method)
        self._registry[method] = ctor

    def create(self, method: str) -> PaymentProcessor:
        ctor = self._registry.get(method)
        if ctor is None:
            raise UnsupportedMethodError("unsupported method " + method)
        return ctor()

# startup: factory.register("UPI", UpiProcessor)   # a class is already a zero-argument callable
```

<div class="lang-note" data-lang="go">

**Go idiom:** most "factories" in Go are just **constructor functions** that return an interface: `func NewUPI(cfg Config) PaymentProcessor`. Only build a registry when the set of types is open-ended. That's how `database/sql` drivers and `image` format decoders work, which is a good example to name in an interview.

</div>

<div class="lang-note" data-lang="python">

**Python idiom:** classes are callables, so a registry is just a `dict` from name to class: `registry["upi"] = UpiProcessor`, then `registry[name]()`. A small decorator can make plugins register themselves (`@register("upi")` above the class), which is a neat Open/Closed point to make. For a fixed, small set of types, a plain `if`/`dict` lookup inside one function is fine. Don't over-build.

</div>

## Builder: many optional parameters, validated once

**Use when** a constructor would take 5+ parameters, many optional, and the object must be **valid and immutable once built**. Typical cases are HTTP requests, search queries, pizzas and notification messages.

```java
final class HttpRequest {
    private final String method, url;
    private final Map<String, String> headers;
    private final Duration timeout;

    private HttpRequest(Builder b) {
        method = b.method; url = b.url; headers = Map.copyOf(b.headers); timeout = b.timeout;
    }

    static Builder builder(String url) { return new Builder(url); }

    static final class Builder {
        private final String url;
        private String method = "GET";
        private final Map<String, String> headers = new LinkedHashMap<>();
        private Duration timeout = Duration.ofSeconds(30);

        private Builder(String url) { this.url = url; }
        Builder method(String m) { this.method = m; return this; }
        Builder header(String k, String v) { headers.put(k, v); return this; }
        Builder timeout(Duration t) { this.timeout = t; return this; }

        HttpRequest build() {
            if (url == null || url.isBlank()) throw new IllegalStateException("url required");
            return new HttpRequest(this);
        }
    }
}
```

In Go, the idiomatic equivalent is **functional options**:

```go
type Request struct {
	method, url string
	headers     map[string]string
	timeout     time.Duration
}

type Option func(*Request)

func WithMethod(m string) Option          { return func(r *Request) { r.method = m } }
func WithHeader(k, v string) Option       { return func(r *Request) { r.headers[k] = v } }
func WithTimeout(d time.Duration) Option  { return func(r *Request) { r.timeout = d } }

func NewRequest(url string, opts ...Option) (*Request, error) {
	r := &Request{method: "GET", url: url, headers: map[string]string{}, timeout: 30 * time.Second}
	for _, o := range opts {
		o(r)
	}
	if url == "" {
		return nil, errors.New("url required")
	}
	return r, nil
}
```

<div class="lang-note" data-lang="python">

In Python, keyword arguments with defaults already solve the "many optional parameters" problem, and a frozen dataclass makes the result immutable. Validate in `__post_init__`. Use an explicit builder (as in the exercise) only when an object is assembled step by step.

```python
@dataclass(frozen=True)
class HttpRequest:
    url: str
    method: str = "GET"
    headers: Mapping[str, str] = field(default_factory=dict)
    timeout: timedelta = timedelta(seconds=30)

    def __post_init__(self) -> None:
        if not self.url:
            raise ValueError("url required")
        object.__setattr__(self, "headers", MappingProxyType(dict(self.headers)))  # defensive, read-only copy


req = HttpRequest("https://api.example.com", method="POST", headers={"x-id": "1"})
```

</div>

**Say:** *"Builder keeps the object immutable and puts validation in one place, in `build()`. Telescoping constructors with nulls are error-prone."*

## Singleton: know it, mostly avoid it

Interviewers ask about Singleton for two reasons: to see if you can write a **thread-safe** one, and to see if you know **why it's usually the wrong choice**.

### The thread-safe implementations

```java
// 1. Holder idiom: lazy, thread-safe via class-loading guarantees. The preferred answer.
final class IdGenerator {
    private final AtomicLong next = new AtomicLong(1);
    private IdGenerator() {}
    private static final class Holder { static final IdGenerator INSTANCE = new IdGenerator(); }
    static IdGenerator getInstance() { return Holder.INSTANCE; }
    long nextId() { return next.getAndIncrement(); }
}

// 2. Double-checked locking: `volatile` is required, or another thread can see a half-built object.
final class Config {
    private static volatile Config instance;
    static Config get() {
        Config c = instance;
        if (c == null) {
            synchronized (Config.class) {
                c = instance;
                if (c == null) instance = c = new Config();
            }
        }
        return c;
    }
}
// 3. enum Singleton { INSTANCE; } is also safe, and survives serialisation and reflection.
```

```go
var (
	instance *IDGenerator
	once     sync.Once
)

func GetIDGenerator() *IDGenerator {
	once.Do(func() { instance = &IDGenerator{} })
	return instance
}

type IDGenerator struct{ next atomic.Int64 }

func (g *IDGenerator) NextID() int64 { return g.next.Add(1) }
```

```python
# 1. The most Pythonic singleton is a module: modules are imported once and cached,
#    so a module-level instance is shared by everyone who imports it.
id_generator = IdGenerator()


# 2. If it must be lazy, use double-checked locking with a lock.
class Config:
    _instance = None
    _lock = threading.Lock()

    @classmethod
    def get(cls) -> "Config":
        if cls._instance is None:
            with cls._lock:
                if cls._instance is None:
                    cls._instance = cls()
        return cls._instance
```

### Why it's usually a red flag

- It's **global mutable state** in disguise: hidden coupling, and tests interfere with each other.
- It **can't be swapped** for a fake. `ParkingLot.getInstance()` inside `Ticket` makes `Ticket` impossible to test on its own.
- "Only one" is a deployment decision, not a property of the class.

**The interview answer:** *"I need a single instance per application, so I'll create it once in `main` and inject it. If you specifically want the Singleton pattern, here's the thread-safe holder version."* This shows you know both the pattern and its trade-off.

Legitimate singletons are stateless or process-wide by nature: a logger facade, a metrics registry, an ID generator.

## Quick picker

| Signal in the problem | Pattern |
|---|---|
| "Support UPI, card, wallet… and more later" | Registry **Factory** behind an interface |
| Object with many optional fields, immutable once built | **Builder** / functional options |
| "There must be exactly one X" | Create once and inject. Explain **Singleton** if asked |
| Copying configured prototypes (game units, documents) | **Prototype**. Mention only |
