---
title: Chain of Responsibility, Template Method, Iterator, Mediator & Visitor
summary: The rest of the behavioral patterns you need, in the order of how often they come up.
---

## Chain of Responsibility: pass it along until someone handles it

**Trigger:** *logging levels and appenders, support-ticket escalation, ATM cash dispensing (2000 → 500 → 100 notes), request validation pipelines, approval workflows (manager → director → VP)*.

Each handler either handles the request, passes it on, or both.

```java
abstract class Approver {
    private Approver next;
    Approver linkWith(Approver next) { this.next = next; return next; }   // fluent chain building
    final String approve(Expense e) {
        if (canApprove(e)) return name();
        if (next == null) throw new IllegalStateException("no one can approve " + e.amount());
        return next.approve(e);
    }
    protected abstract boolean canApprove(Expense e);
    protected abstract String name();
}

// manager.linkWith(director).linkWith(vp);  manager.approve(expense);
```

```go
type Handler interface {
	Handle(r Request) (handled bool)
}

// In Go a chain is often just a slice, which is simpler than linked handlers:
type Chain []Handler

func (c Chain) Handle(r Request) bool {
	for _, h := range c {
		if h.Handle(r) {
			return true
		}
	}
	return false
}
```

```python
class Approver(ABC):
    def __init__(self) -> None:
        self._next: Optional[Approver] = None

    def link_with(self, nxt: "Approver") -> "Approver":  # fluent chain building
        self._next = nxt
        return nxt

    def approve(self, e: Expense) -> str:
        if self.can_approve(e):
            return self.name()
        if self._next is None:
            raise ApprovalError("no one can approve %d" % e.amount)
        return self._next.approve(e)

    @abstractmethod
    def can_approve(self, e: Expense) -> bool: ...

    @abstractmethod
    def name(self) -> str: ...

# manager.link_with(director).link_with(vp);  manager.approve(expense)
```

**ATM dispense variant** (asked often): each handler dispenses as many notes of its denomination as possible and passes the **remainder** along. If anything is left at the end, fail and roll back.

**Two styles to distinguish out loud:**

- *Stop at the first handler* (approvals, ticket escalation)
- *Every handler gets a look* (logging appenders, middleware). Here each one decides independently, and the chain always continues

## Template Method: fixed skeleton, variable steps

**Trigger:** several processes share the same steps in the same order but differ in some of them. Report generation, payment flows (validate → authorise → capture → notify), game turns, data importers.

```java
abstract class DataImporter {
    public final ImportResult run(Path file) {          // final: the skeleton can't be changed
        List<String> rows = read(file);
        List<Record> parsed = rows.stream().map(this::parse).toList();
        validate(parsed);
        return save(parsed);
    }
    protected abstract Record parse(String row);       // varies: CSV vs JSON lines
    protected void validate(List<Record> rs) {}        // optional hook
    private List<String> read(Path p) { /* shared */ }
    private ImportResult save(List<Record> rs) { /* shared */ }
}
```

```python
class DataImporter(ABC):
    def run(self, path: Path) -> ImportResult:  # the skeleton: subclasses don't override this
        rows = self._read(path)
        parsed = [self.parse(r) for r in rows]
        self.validate(parsed)
        return self._save(parsed)

    @abstractmethod
    def parse(self, row: str) -> Record: ...  # varies: CSV vs JSON lines

    def validate(self, records: List[Record]) -> None:  # optional hook
        pass

    def _read(self, path: Path) -> List[str]: ...  # shared
    def _save(self, records: List[Record]) -> ImportResult: ...  # shared
```

In Go there are no abstract classes. Pass the varying steps as an **interface or function fields** into a struct that owns the skeleton. That's Strategy, and Go programmers will expect you to say so.

## Iterator: traverse without exposing the internals

You use it every day (`for (x : list)`, `range`). In LLD, it comes up when you build a **custom collection**, such as a playlist with shuffle or repeat modes, a paginated API client, or tree traversal. Say *"I'll expose an `Iterator<Song>` so callers don't depend on how the playlist stores songs."*

Go 1.23+ has range-over-func iterators (`iter.Seq[T]`), and Python has generators. Both make this idiomatic:

```go
func (p *Playlist) All() iter.Seq[Song] {
	return func(yield func(Song) bool) {
		for _, i := range p.order {
			if !yield(p.songs[i]) {
				return
			}
		}
	}
}
// for s := range playlist.All() { ... }
```

```python
class Playlist:
    def __iter__(self) -> Iterator[Song]:  # a generator: `yield` makes it lazy, the storage stays private
        for i in self._order:
            yield self._songs[i]

# for song in playlist: ...
```

## Mediator: objects talk through a hub, not to each other

**Trigger:** *chat rooms, air-traffic control, auction houses, an elevator controller coordinating many elevators*. Without a mediator, every component references every other one (N² links). With it, every component knows only the mediator.

```java
interface ChatRoom { void send(String from, String to, String msg); void join(User u); }
// Users call room.send(...) and never hold references to other users.
```

```python
class ChatRoom(ABC):
    @abstractmethod
    def send(self, sender: str, to: str, msg: str) -> None: ...

    @abstractmethod
    def join(self, user: User) -> None: ...

# Users call room.send(...) and never hold references to other users.
```

The **elevator dispatcher** you'll build on Day 8 is a mediator: requests go to the controller, and it decides which car serves them.

## Visitor: add operations to a stable class hierarchy

**Trigger:** you have a **fixed** set of node types (an AST, shapes, document elements) and keep adding **operations** (render, export to JSON, compute area, tax). Visitor puts each operation in its own class using double dispatch.

Use it sparingly. It's rarely the right answer in a 45-minute LLD round, and it's heavy. Mention it if the interviewer asks *"how would you add exporting without touching every shape class?"* In modern Java, a `sealed interface` plus a pattern-matching `switch` often replaces it.

## Memento: snapshots for undo

Capture an object's state in an opaque snapshot, then restore it later. It's the alternative to Command-based undo (see the Command lesson). Typical uses are game save points, form drafts and editor checkpoints.

## Summary table

| Pattern | One-line intent | Classic LLD problem |
|---|---|---|
| Chain of Responsibility | Pass a request along handlers | Logger, ATM dispenser, approvals |
| Template Method | Fixed algorithm, variable steps | Payment flow, report export |
| Iterator | Traverse without exposing the structure | Playlist, paginated results |
| Mediator | Central hub coordinating peers | Chat room, elevator controller |
| Visitor | New operations over stable types | Expression evaluator, exporters |
| Memento | Snapshot and restore state | Undo in games or editors |
