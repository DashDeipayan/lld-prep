Implement the part of chess that interviewers ask you to code: **pseudo-legal move generation**, using polymorphic pieces, plus a `move` that respects turns. Check, castling, en passant and promotion are **out of scope**. Read the chess walkthrough lesson first.

## Board and notation

Squares use algebraic notation: files `a`–`h`, ranks `1`–`8`. White starts on ranks 1–2. `pieceAt(sq)` returns a colour letter plus a piece letter (`"wK"`, `"bN"`, `"wP"`, …) or `"."` for an empty square. Piece letters: **K** king, **Q** queen, **R** rook, **B** bishop, **N** knight, **P** pawn.

## Movement rules (pseudo-legal)

| Piece | Moves |
|---|---|
| King | one square in any of the 8 directions |
| Queen | any distance along ranks, files and diagonals |
| Rook | any distance along ranks and files |
| Bishop | any distance along diagonals |
| Knight | the 8 L-shaped jumps. **Ignores blockers** |
| Pawn | 1 forward if empty. 2 forward from its start rank if both squares are empty. Captures 1 diagonally forward **only** if an enemy piece is there |

Sliding pieces stop at the first occupied square. They **include** it if it's an enemy (a capture) and **exclude** it if it's a friendly piece. Nothing moves onto a friendly piece.

## API

| Java (`ChessBoard`) | Go (`*Board`) |
|---|---|
| `ChessBoard.empty()`, `ChessBoard.standard()` | `NewEmptyBoard()`, `NewStandardBoard()` |
| `place(sq, color, kind)` | `Place(sq, c, k) error` |
| `pieceAt(sq)` → `"wK"` / `"."` | `PieceAt(sq) string` |
| `movesFrom(sq)` → target squares **sorted** (empty list for an empty square) | `MovesFrom(sq) ([]string, error)` |
| `move(from, to)` | `Move(from, to) error` |
| `turn()` → `WHITE` first | `Turn() Color` |

`move` errors: no piece at `from`, or `to` isn't a pseudo-legal target → *illegal move*. The piece isn't the side to move → *wrong turn*. A malformed square → *invalid square*.

<div class="lang-note" data-lang="java">

Java: `enum Color { WHITE, BLACK }` and `enum Kind { KING, QUEEN, ROOK, BISHOP, KNIGHT, PAWN }` are provided. Illegal move or invalid square throws `IllegalArgumentException`. Wrong turn throws `IllegalStateException`.

</div>

<div class="lang-note" data-lang="go">

Go: `Color` (`White`, `Black`) and `Kind` (`King`, `Queen`, `Rook`, `Bishop`, `Knight`, `Pawn`) are provided. Errors: `ErrIllegalMove`, `ErrWrongTurn`, `ErrInvalidSquare`.

</div>
