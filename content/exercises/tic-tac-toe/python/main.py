from solution import Player, TicTacToe

g = TicTacToe(3)
p = Player.X
for r, c in [(0, 0), (1, 1), (0, 1), (2, 2), (0, 2)]:
    print("%s -> (%d,%d): %s" % (p.value, r, c, g.move(p, r, c).value))
    p = Player.O if p is Player.X else Player.X
for r in range(3):
    print("".join(g.cell(r, c) for c in range(3)))
