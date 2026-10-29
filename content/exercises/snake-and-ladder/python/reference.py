from __future__ import annotations

from dataclasses import dataclass
from typing import Callable, Dict, List, Optional

# Dice: any zero-argument callable returning 1..6. Tests inject a scripted one.
Dice = Callable[[], int]


@dataclass(frozen=True)
class TurnResult:
    player: str
    roll: int
    from_pos: int  # `from` is a Python keyword
    to_pos: int
    won: bool


class GameOverError(Exception):
    """Provided: play_turn() after somebody has won."""


class Board:
    def __init__(self, size: int, snakes: Dict[int, int], ladders: Dict[int, int]) -> None:
        if size < 2:
            raise ValueError("size too small")
        self.size = size
        self._jumps: Dict[int, int] = {}
        for head, tail in snakes.items():
            self._add(head, tail, head > tail, "snake")
        for bottom, top in ladders.items():
            self._add(bottom, top, bottom < top, "ladder")
        for end in self._jumps.values():
            if end in self._jumps:  # no chains, so no cycles
                raise ValueError("jump ends on another jump's start: %d" % end)

    def _add(self, start: int, end: int, right_direction: bool, kind: str) -> None:
        if not right_direction:
            raise ValueError("%s goes the wrong way: %d->%d" % (kind, start, end))
        if not (1 <= start < self.size and 1 <= end <= self.size):
            raise ValueError(kind + " out of range")
        if start in self._jumps:
            raise ValueError("two jumps start at %d" % start)
        self._jumps[start] = end

    def resolve(self, cell: int) -> int:
        """Where a player who lands on `cell` ends up."""
        return self._jumps.get(cell, cell)


class Game:
    def __init__(self, board: Board, players: List[str], dice: Dice) -> None:
        if len(players) < 2:
            raise ValueError("need at least 2 players")
        self._board = board
        self._players = tuple(players)
        self._dice = dice
        self._pos: Dict[str, int] = {p: 0 for p in players}
        self._current = 0
        self._sixes = 0
        self._turn_start = 0
        self._winner: Optional[str] = None

    def play_turn(self) -> TurnResult:
        if self._winner is not None:
            raise GameOverError("game over")
        p = self._players[self._current]
        start = self._pos[p]
        if self._sixes == 0:
            self._turn_start = start
        roll = self._dice()

        if roll == 6:
            self._sixes += 1
            if self._sixes == 3:  # penalty: undo the whole turn
                self._pos[p] = self._turn_start
                self._pass_turn()
                return TurnResult(p, roll, start, self._turn_start, False)

        target = start + roll
        end = start if target > self._board.size else self._board.resolve(target)
        self._pos[p] = end

        if end == self._board.size:
            self._winner = p
            return TurnResult(p, roll, start, end, True)
        if roll != 6:
            self._pass_turn()
        return TurnResult(p, roll, start, end, False)

    def _pass_turn(self) -> None:
        self._current = (self._current + 1) % len(self._players)
        self._sixes = 0

    def position(self, player: str) -> int:
        return self._pos[player]

    def winner(self) -> Optional[str]:
        return self._winner
