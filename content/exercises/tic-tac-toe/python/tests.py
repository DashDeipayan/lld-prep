from lldtest import eq, raises
from solution import GameStateError, InvalidMoveError, Player, Status, TicTacToe


def play(g, moves):
    p, s = Player.X, Status.IN_PROGRESS
    for r, c in moves:
        s = g.move(p, r, c)
        p = Player.O if p is Player.X else Player.X
    return s


def test_row_win():
    g = TicTacToe(3)
    eq(Status.X_WON, play(g, [(0, 0), (1, 0), (0, 1), (1, 1), (0, 2)]), "X wins top row")
    eq("X", g.cell(0, 2), "cell")
    eq(".", g.cell(2, 2), "empty cell")


def test_column_and_diagonal_wins():
    eq(Status.O_WON, play(TicTacToe(3), [(0, 0), (0, 1), (1, 0), (1, 1), (2, 2), (2, 1)]), "O wins middle column")
    eq(Status.X_WON, play(TicTacToe(3), [(0, 0), (0, 1), (1, 1), (0, 2), (2, 2)]), "X main diagonal")
    eq(Status.X_WON, play(TicTacToe(3), [(0, 2), (0, 0), (1, 1), (0, 1), (2, 0)]), "X anti-diagonal")


def test_larger_board():
    s = play(TicTacToe(4), [(0, 3), (0, 0), (1, 2), (1, 0), (2, 1), (2, 0), (3, 0)])
    eq(Status.X_WON, s, "4x4 anti-diagonal completed by X at (3,0)")


def test_draw():
    s = play(TicTacToe(3), [(0, 0), (0, 1), (0, 2), (1, 1), (1, 0), (1, 2), (2, 1), (2, 0), (2, 2)])
    eq(Status.DRAW, s, "full board, no line")


def test_invalid_moves_change_nothing():
    g = TicTacToe(3)
    raises(GameStateError, lambda: g.move(Player.O, 0, 0), "X moves first")
    g.move(Player.X, 1, 1)
    raises(InvalidMoveError, lambda: g.move(Player.O, 1, 1), "occupied")
    raises(InvalidMoveError, lambda: g.move(Player.O, 3, 0), "out of bounds")
    raises(GameStateError, lambda: g.move(Player.X, 0, 0), "X can't move twice")
    eq(Status.IN_PROGRESS, g.move(Player.O, 0, 0), "O's valid move still accepted")
    raises(ValueError, lambda: TicTacToe(2), "n >= 3")


def test_no_moves_after_game_over():
    g = TicTacToe(3)
    play(g, [(0, 0), (1, 0), (0, 1), (1, 1), (0, 2)])
    raises(GameStateError, lambda: g.move(Player.O, 2, 2), "game over")
    eq(Status.X_WON, g.status(), "status sticks")
