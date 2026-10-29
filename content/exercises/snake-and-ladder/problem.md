A staple of machine-coding rounds (Flipkart, Swiggy, Uber India). Most candidates can make it run. What scores is **clean rule handling** and **testability**.

## Board

`Board(size, snakes, ladders)`: cells are numbered `1..size`. `snakes` maps head → tail (head > tail), `ladders` maps bottom → top (bottom < top). The board is **invalid** if:

- any endpoint is outside `1..size`, or a jump starts on `size`
- a snake goes up or a ladder goes down (or start == end)
- two jumps start on the same cell
- a jump **ends on a cell where another jump starts** (no chains, so no cycles)

## Game

`Game(board, players, dice)`: at least 2 players. Everyone starts at position **0**, off the board. `playTurn()` performs **one roll** for the current player and returns `TurnResult(player, roll, from, to, won)`:

1. `target = from + roll`. If `target > size`, the player **stays** (`to == from`).
2. Otherwise move to `target`. If a snake or ladder starts there, follow it (once).
3. If `to == size`, that player **wins** and the game is over.
4. **Next player:** rolling a **6** gives the **same** player another roll. Their **third consecutive six** is a penalty: they go back to where they were at the **start of the turn** (`to` = that position), and play passes to the next player.

`position(player)`, `winner()`, and `playTurn()` after the game is over (an error).

## Example

Size 20, ladder 3→11, snake 17→5, players A and B, dice `2, 4, 6, 1, …`

```
A rolls 2: 0 → 2
B rolls 4: 0 → 4
A rolls 6: 2 → 8      (six: A rolls again)
A rolls 1: 8 → 9
```

<div class="lang-note" data-lang="java">

Java: `Dice` is an interface (`int roll()`), and `TurnResult` and `Board`'s constructor signature are provided. An invalid board or fewer than 2 players throws `IllegalArgumentException`. A turn after game over throws `IllegalStateException`. `winner()` returns `Optional<String>`.

</div>

<div class="lang-note" data-lang="go">

Go: `NewBoard(size int, snakes, ladders map[int]int) (*Board, error)` (`ErrInvalidBoard`). `NewGame(b *Board, players []string, d Dice) (*Game, error)` (`ErrNotEnoughPlayers`). `PlayTurn() (TurnResult, error)` (`ErrGameOver`). `Position(p string) int`, `Winner() (string, bool)`.

</div>

<div class="lang-note" data-lang="python">

Python: dice is any zero-argument callable returning 1–6. `TurnResult` is a frozen dataclass with fields `player, roll, from_pos, to_pos, won` (`from` is a keyword in Python). An invalid board or fewer than 2 players raises `ValueError`. A turn after game over raises the provided `GameOverError`. `Game` has `play_turn()`, `position(player)`, and `winner()`, which returns the name or `None`.

</div>
