Design an N×N tic-tac-toe game for two players, `X` and `O`. **X moves first.** A player wins by filling an entire **row, column or diagonal**. When the board is full and nobody has won, it's a **draw**.

The interviewer's twist: *"`move` must be O(1). Don't scan the board."*

## API

| Java (`TicTacToe`) | Go (`*Game`) | Behaviour |
|---|---|---|
| `new TicTacToe(n)` | `NewGame(n)` | `n ≥ 3` |
| `move(player, row, col)` → `Status` | `Move(p, row, col) (Status, error)` | place a mark and return the status **after** the move |
| `status()` | `Status()` | `IN_PROGRESS`, `X_WON`, `O_WON` or `DRAW` |
| `cell(row, col)` → `'X'`, `'O'` or `'.'` | `Cell(r, c) byte` | for display and tests |

Errors (nothing changes when a move is rejected):

- the game is already over
- row or column out of bounds
- the cell is occupied
- it isn't that player's turn

<div class="lang-note" data-lang="java">

Java: `enum Player { X, O }`, `enum Status { IN_PROGRESS, X_WON, O_WON, DRAW }` (provided). Game over and wrong turn throw `IllegalStateException`. Bounds and occupied cell throw `IllegalArgumentException`. `n < 3` throws `IllegalArgumentException`.

</div>

<div class="lang-note" data-lang="go">

Go: `Player` (`X`, `O`) and `Status` (`InProgress`, `XWon`, `OWon`, `Draw`) are provided. Errors: `ErrGameOver`, `ErrOutOfBounds`, `ErrOccupied`, `ErrWrongTurn`.

</div>

<div class="lang-note" data-lang="python">

Python: the `Player` and `Status` enums are provided, along with two errors: bounds and occupied cells raise `InvalidMoveError`, and game over and wrong turn raise `GameStateError`. `TicTacToe(n)` has `move(player, row, col)` (which returns the `Status`), `status()` and `cell(row, col)`. `n < 3` raises `ValueError`.

</div>

## Extension questions

- *K-in-a-row on a large board (Gomoku)?* (Counters no longer work. Scan the 4 directions from the last move, which is O(k).)
- *Human vs AI?* (Use a `PlayerStrategy` interface with `nextMove(board)`, and implement `HumanPlayer` and `MinimaxPlayer`.)
- *Undo?* (Command: store the moves and reverse the counter deltas.)
