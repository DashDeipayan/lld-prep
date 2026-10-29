import random

from solution import Board, Game

board = Board(100, {99: 7, 62: 19, 48: 26}, {2: 38, 28: 84, 51: 67, 71: 91})
rnd = random.Random(42)
g = Game(board, ["alice", "bob"], lambda: rnd.randint(1, 6))
turns = 0
while g.winner() is None and turns < 1000:
    r = g.play_turn()
    if turns < 8:
        print(r)
    turns += 1
print("... winner =", g.winner() or "none", "after", turns, "rolls")
