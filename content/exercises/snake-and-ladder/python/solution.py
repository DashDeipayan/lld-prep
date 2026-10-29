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
        pass  # TODO: validate and store


class Game:
    def __init__(self, board: Board, players: List[str], dice: Dice) -> None:
        pass  # TODO

    def play_turn(self) -> TurnResult:
        raise NotImplementedError("TODO")

    def position(self, player: str) -> int:
        raise NotImplementedError("TODO")

    def winner(self) -> Optional[str]:
        raise NotImplementedError("TODO")
