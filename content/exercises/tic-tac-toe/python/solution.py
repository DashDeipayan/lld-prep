from __future__ import annotations

from enum import Enum


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
        pass  # TODO

    def move(self, player: Player, row: int, col: int) -> Status:
        raise NotImplementedError("TODO")

    def status(self) -> Status:
        raise NotImplementedError("TODO")

    def cell(self, row: int, col: int) -> str:
        """'X', 'O' or '.'"""
        raise NotImplementedError("TODO")
