Implement the core of a text editor where every edit can be undone and redone. Positions are character indices into the current text. Use ASCII only, no Unicode subtleties.

## API

| Java | Go | Behaviour |
|---|---|---|
| `new TextEditor(maxHistory)` | `NewTextEditor(maxHistory)` | empty document. Keeps at most `maxHistory` undoable edits |
| `text()` | `Text()` | current content |
| `insert(pos, s)` | `Insert(pos, s) error` | insert `s` at `pos` (`0 ≤ pos ≤ len`) |
| `delete(from, to)` | `Delete(from, to) error` | remove `[from, to)` (`0 ≤ from ≤ to ≤ len`) |
| `undo()` | `Undo() bool` | revert the most recent edit. `false` if there's nothing to undo |
| `redo()` | `Redo() bool` | re-apply the most recently undone edit. `false` if there's nothing to redo |

## Rules

1. Invalid positions are an error (Java `IndexOutOfBoundsException`, Go `ErrOutOfRange`) and **change nothing**, history included.
2. No-ops (inserting `""`, or deleting an empty range) are accepted but **not recorded**.
3. A new edit after an undo **clears the redo history**.
4. When more than `maxHistory` edits are recorded, the **oldest** one can no longer be undone.

## Example

```
insert(0,"hello")        → "hello"
insert(5," world")       → "hello world"
delete(0,6)              → "world"
undo()                   → "hello world"
undo()                   → "hello"
redo()                   → "hello world"
insert(11,"!")           → "hello world!"   (redo history cleared)
redo()                   → false
```

## Discussion

- How would you implement **grouping** ("undo the last word typed" instead of each character)? (A `MacroCommand` that holds a list.)
- When would **Memento** (snapshotting the whole text) beat Command? (Small documents, or complex edits with hard-to-write inverses.)
