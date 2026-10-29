from lldtest import eq, raises
from solution import Board, Game, GameOverError, TurnResult


def script(*rolls):
    it = iter(rolls)
    return lambda: next(it)


def board():
    return Board(20, {17: 5}, {3: 11})


def test_board_validation():
    raises(ValueError, lambda: Board(20, {5: 9}, {}), "snake going up")
    raises(ValueError, lambda: Board(20, {}, {9: 5}), "ladder going down")
    raises(ValueError, lambda: Board(20, {25: 3}, {}), "out of range")
    raises(ValueError, lambda: Board(20, {20: 3}, {}), "jump from final cell")
    raises(ValueError, lambda: Board(20, {9: 2}, {9: 15}), "two jumps from same cell")
    raises(ValueError, lambda: Board(20, {15: 4}, {4: 12}), "chain: ladder 4->12, snake 15->4 ends on 4")
    raises(ValueError, lambda: Game(board(), ["solo"], script(1)), "need 2 players")


def test_basic_turns_alternate():
    g = Game(board(), ["A", "B"], script(2, 4, 1))
    eq(TurnResult("A", 2, 0, 2, False), g.play_turn(), "A")
    eq(TurnResult("B", 4, 0, 4, False), g.play_turn(), "B")
    eq(TurnResult("A", 1, 2, 11, False), g.play_turn(), "A lands on ladder 3 -> 11")
    eq(11, g.position("A"), "position")


def test_snake():
    g = Game(board(), ["A", "B"], script(5, 1, 6, 6))
    g.play_turn()
    g.play_turn()
    eq(TurnResult("A", 6, 5, 11, False), g.play_turn(), "A: 5 -> 11 (six, rolls again)")
    eq(TurnResult("A", 6, 11, 5, False), g.play_turn(), "A: 11 -> 17, snake down to 5")


def test_six_gives_extra_roll():
    g = Game(board(), ["A", "B"], script(6, 1, 2))
    g.play_turn()
    eq("A", g.play_turn().player, "A rolls again after a six")
    eq("B", g.play_turn().player, "then B")


def test_three_sixes_penalty():
    g = Game(Board(40, {}, {}), ["A", "B"], script(2, 1, 6, 6, 6, 3))
    for _ in range(4):
        g.play_turn()
    third = g.play_turn()
    eq(TurnResult("A", 6, 14, 2, False), third, "third six sends A back to the start of the turn")
    eq(2, g.position("A"), "A back at 2")
    eq("B", g.play_turn().player, "turn passes after the penalty")


def test_exact_finish_and_winner():
    g = Game(Board(10, {}, {}), ["A", "B"], script(5, 1, 6, 1, 4, 4))
    g.play_turn()
    g.play_turn()
    eq(TurnResult("A", 6, 5, 5, False), g.play_turn(), "overshoot: 5+6 > 10, stay (six still grants a reroll)")
    eq(TurnResult("A", 1, 5, 6, False), g.play_turn(), "A again")
    g.play_turn()
    win = g.play_turn()
    eq(TurnResult("A", 4, 6, 10, True), win, "A lands exactly on 10")
    eq("A", g.winner(), "winner")
    raises(GameOverError, g.play_turn, "no turns after game over")
