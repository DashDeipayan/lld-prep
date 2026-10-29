from __future__ import annotations

from enum import Enum
from typing import List, Optional


class Player(Enum):
    X = "X"
    O = "O"


class Status(Enum):
    IN_PROGRESS = "IN_PROGRESS"
    X_WON = "X_WON"
    O_WON = "O_WON"
    DRAW = "DRAW"


class InvalidMoveError(ValueError):
    """Provided: out of bounds, or the cell is occupied."""


class GameStateError(Exception):
    """Provided: the game is over, or it isn't that player's turn."""


class TicTacToe:
    def __init__(self, n: int) -> None:
        if n < 3:
            raise ValueError("n must be >= 3")
        self._n = n
        self._board: List[List[Optional[Player]]] = [[None] * n for _ in range(n)]
        # O(1) win check: X adds +1, O adds -1; a line is won when its sum reaches ±n.
        self._rows = [0] * n
        self._cols = [0] * n
        self._diag = 0
        self._anti = 0
        self._moves = 0
        self._turn = Player.X
        self._status = Status.IN_PROGRESS

    def move(self, player: Player, row: int, col: int) -> Status:
        # Validate everything before changing anything.
        if self._status is not Status.IN_PROGRESS:
            raise GameStateError("game over: " + self._status.value)
        if not (0 <= row < self._n and 0 <= col < self._n):
            raise InvalidMoveError("out of bounds")
        if self._board[row][col] is not None:
            raise InvalidMoveError("cell occupied")
        if player is not self._turn:
            raise GameStateError("not %s's turn" % player.value)

        self._board[row][col] = player
        self._moves += 1
        d = 1 if player is Player.X else -1
        self._rows[row] += d
        self._cols[col] += d
        if row == col:
            self._diag += d
        if row + col == self._n - 1:
            self._anti += d

        target = d * self._n
        if target in (self._rows[row], self._cols[col], self._diag, self._anti):
            self._status = Status.X_WON if player is Player.X else Status.O_WON
        elif self._moves == self._n * self._n:
            self._status = Status.DRAW
        self._turn = Player.O if self._turn is Player.X else Player.X
        return self._status

    def status(self) -> Status:
        return self._status

    def cell(self, row: int, col: int) -> str:
        p = self._board[row][col]
        return "." if p is None else p.value
