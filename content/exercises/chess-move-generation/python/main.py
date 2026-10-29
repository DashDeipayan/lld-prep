from solution import ChessBoard

b = ChessBoard.standard()
print("b1 knight:", b.moves_from("b1"))
print("e2 pawn:  ", b.moves_from("e2"))
b.move("e2", "e4")
b.move("d7", "d5")
print("e4 pawn can capture:", b.moves_from("e4"))
b.move("e4", "d5")
print("d5 now holds", b.piece_at("d5"), ", turn =", b.turn().name)
print("d1 queen: ", b.moves_from("d1"))
