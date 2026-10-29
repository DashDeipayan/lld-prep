from __future__ import annotations

from abc import ABC, abstractmethod
from enum import Enum
from typing import Dict, List, NamedTuple, Optional, Sequence, Tuple


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


class Square(NamedTuple):
    file: int  # 0..7 = a..h
    rank: int  # 0..7 = 1..8

    @staticmethod
    def parse(s: str) -> "Square":
        if not isinstance(s, str) or len(s) != 2 or s[0] not in "abcdefgh" or s[1] not in "12345678":
            raise ValueError("invalid square %r" % s)
        return Square(ord(s[0]) - ord("a"), int(s[1]) - 1)

    def offset(self, df: int, dr: int) -> Optional["Square"]:
        """None when off the board."""
        f, r = self.file + df, self.rank + dr
        return Square(f, r) if 0 <= f < 8 and 0 <= r < 8 else None

    def __str__(self) -> str:
        return "abcdefgh"[self.file] + str(self.rank + 1)


Dirs = Sequence[Tuple[int, int]]
STRAIGHT: Dirs = ((1, 0), (-1, 0), (0, 1), (0, -1))
DIAGONAL: Dirs = ((1, 1), (1, -1), (-1, 1), (-1, -1))
ALL: Dirs = STRAIGHT + DIAGONAL
KNIGHT: Dirs = ((1, 2), (2, 1), (2, -1), (1, -2), (-1, -2), (-2, -1), (-2, 1), (-1, 2))


class Piece(ABC):
    def __init__(self, color: Color, kind: Kind) -> None:
        self.color = color
        self.kind = kind

    @abstractmethod
    def candidate_moves(self, board: "ChessBoard", src: Square) -> List[Square]: ...

    def _walk(self, board: "ChessBoard", src: Square, dirs: Dirs, max_range: int) -> List[Square]:
        """Walk each direction up to max_range squares; stop at the first piece, including it if it's an enemy."""
        out = []
        for df, dr in dirs:
            sq = src
            for _ in range(max_range):
                sq = sq.offset(df, dr)
                if sq is None:
                    break
                other = board.at(sq)
                if other is None:
                    out.append(sq)
                    continue
                if other.color is not self.color:
                    out.append(sq)
                break
        return out

    @staticmethod
    def of(color: Color, kind: Kind) -> "Piece":
        if kind is Kind.PAWN:
            return Pawn(color)
        dirs, max_range = {
            Kind.KING: (ALL, 1), Kind.QUEEN: (ALL, 7), Kind.ROOK: (STRAIGHT, 7),
            Kind.BISHOP: (DIAGONAL, 7), Kind.KNIGHT: (KNIGHT, 1),
        }[kind]
        return Stepper(color, kind, dirs, max_range)


class Stepper(Piece):
    """King, queen, rook, bishop and knight differ only in directions and range."""

    def __init__(self, color: Color, kind: Kind, dirs: Dirs, max_range: int) -> None:
        super().__init__(color, kind)
        self._dirs = dirs
        self._range = max_range

    def candidate_moves(self, board, src):
        return self._walk(board, src, self._dirs, self._range)


class Pawn(Piece):
    def __init__(self, color: Color) -> None:
        super().__init__(color, Kind.PAWN)

    def candidate_moves(self, board, src):
        out = []
        step = 1 if self.color is Color.WHITE else -1
        start_rank = 1 if self.color is Color.WHITE else 6
        one = src.offset(0, step)
        if one is not None and board.at(one) is None:
            out.append(one)
            two = one.offset(0, step)
            if src.rank == start_rank and two is not None and board.at(two) is None:
                out.append(two)
        for df in (-1, 1):
            cap = src.offset(df, step)
            if cap is not None and board.at(cap) is not None and board.at(cap).color is not self.color:
                out.append(cap)
        return out


class ChessBoard:
    def __init__(self) -> None:
        self._grid: Dict[Square, Piece] = {}
        self._turn = Color.WHITE

    @classmethod
    def empty(cls) -> "ChessBoard":
        return cls()

    @classmethod
    def standard(cls) -> "ChessBoard":
        b = cls()
        back = [Kind.ROOK, Kind.KNIGHT, Kind.BISHOP, Kind.QUEEN, Kind.KING, Kind.BISHOP, Kind.KNIGHT, Kind.ROOK]
        for f, kind in enumerate(back):
            b._grid[Square(f, 0)] = Piece.of(Color.WHITE, kind)
            b._grid[Square(f, 1)] = Piece.of(Color.WHITE, Kind.PAWN)
            b._grid[Square(f, 6)] = Piece.of(Color.BLACK, Kind.PAWN)
            b._grid[Square(f, 7)] = Piece.of(Color.BLACK, kind)
        return b

    def at(self, sq: Square) -> Optional[Piece]:
        return self._grid.get(sq)

    def place(self, square: str, color: Color, kind: Kind) -> None:
        self._grid[Square.parse(square)] = Piece.of(color, kind)

    def piece_at(self, square: str) -> str:
        p = self.at(Square.parse(square))
        return "." if p is None else p.color.value + p.kind.value

    def moves_from(self, square: str) -> List[str]:
        sq = Square.parse(square)
        p = self.at(sq)
        return [] if p is None else sorted(str(t) for t in p.candidate_moves(self, sq))

    def move(self, src: str, dst: str) -> None:
        s, d = Square.parse(src), Square.parse(dst)
        p = self.at(s)
        if p is None:
            raise IllegalMoveError("no piece at " + src)
        if p.color is not self._turn:
            raise WrongTurnError("it is %s's turn" % self._turn.name)
        if d not in p.candidate_moves(self, s):
            raise IllegalMoveError("illegal move %s-%s" % (src, dst))
        self._grid[d] = p
        del self._grid[s]
        self._turn = Color.BLACK if self._turn is Color.WHITE else Color.WHITE

    def turn(self) -> Color:
        return self._turn
