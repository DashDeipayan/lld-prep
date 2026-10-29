from __future__ import annotations

from enum import Enum
from typing import List


class Color(Enum):
    WHITE = "w"
    BLACK = "b"


class Kind(Enum):
    KING = "K"
    QUEEN = "Q"
    ROOK = "R"
    BISHOP = "B"
    KNIGHT = "N"
    PAWN = "P"


class IllegalMoveError(ValueError):
    """Provided: no piece at the source, or the target isn't a pseudo-legal move."""


class WrongTurnError(Exception):
    """Provided: the piece doesn't belong to the side to move."""


# TODO: a Square value type, an abstract Piece with one subclass per kind (or a data-driven design)


class ChessBoard:
    @classmethod
    def empty(cls) -> "ChessBoard":
        raise NotImplementedError("TODO")

    @classmethod
    def standard(cls) -> "ChessBoard":
        raise NotImplementedError("TODO")

    def place(self, square: str, color: Color, kind: Kind) -> None:
        raise NotImplementedError("TODO")

    def piece_at(self, square: str) -> str:
        """"wK", "bN", ... or "." for an empty square."""
        raise NotImplementedError("TODO")

    def moves_from(self, square: str) -> List[str]:
        raise NotImplementedError("TODO")

    def move(self, src: str, dst: str) -> None:
        raise NotImplementedError("TODO")

    def turn(self) -> Color:
        raise NotImplementedError("TODO")
