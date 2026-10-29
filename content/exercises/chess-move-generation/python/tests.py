from lldtest import check, eq, raises
from solution import ChessBoard, Color, IllegalMoveError, Kind, WrongTurnError


def test_standard_opening_moves():
    b = ChessBoard.standard()
    eq(["a3", "c3"], b.moves_from("b1"), "knight jumps over pawns")
    eq(["e3", "e4"], b.moves_from("e2"), "pawn single + double step")
    eq([], b.moves_from("a1"), "rook boxed in")
    eq(["f6", "h6"], b.moves_from("g8"), "black knight")
    eq(["d5", "d6"], b.moves_from("d7"), "black pawn moves down")
    eq([], b.moves_from("e5"), "empty square")
    eq("wK", b.piece_at("e1"), "white king")
    eq("bN", b.piece_at("b8"), "black knight letter is N")


def test_sliders_on_empty_board():
    b = ChessBoard.empty()
    b.place("d4", Color.WHITE, Kind.ROOK)
    eq(14, len(b.moves_from("d4")), "rook in the middle")
    b.place("a1", Color.WHITE, Kind.BISHOP)
    eq(["b2", "c3"], b.moves_from("a1"), "bishop stops before its own rook on d4")
    b.place("h8", Color.BLACK, Kind.QUEEN)
    rank_and_file = [s for s in b.moves_from("h8") if s.endswith("8") or s.startswith("h")]
    eq(["a8", "b8", "c8", "d8", "e8", "f8", "g8", "h1", "h2", "h3", "h4", "h5", "h6", "h7"], rank_and_file, "queen rank + file")
    check("d4" in b.moves_from("h8"), "queen diagonal reaches d4 (capture)")
    check("c3" not in b.moves_from("h8"), "queen stops at the captured piece")


def test_blocking_and_captures():
    b = ChessBoard.empty()
    b.place("a1", Color.WHITE, Kind.ROOK)
    b.place("a4", Color.BLACK, Kind.PAWN)
    b.place("c1", Color.WHITE, Kind.KNIGHT)
    eq(["a2", "a3", "a4", "b1"], b.moves_from("a1"), "capture a4, blocked by own knight on c1")


def test_king_and_knight():
    b = ChessBoard.empty()
    b.place("e4", Color.WHITE, Kind.KING)
    eq(["d3", "d4", "d5", "e3", "e5", "f3", "f4", "f5"], b.moves_from("e4"), "king in the centre")
    b.place("a1", Color.BLACK, Kind.KING)
    eq(["a2", "b1", "b2"], b.moves_from("a1"), "king in the corner")
    b.place("g1", Color.WHITE, Kind.KNIGHT)
    eq(["e2", "f3", "h3"], b.moves_from("g1"), "knight near the edge")


def test_pawn_rules():
    b = ChessBoard.empty()
    b.place("e2", Color.WHITE, Kind.PAWN)
    b.place("e3", Color.BLACK, Kind.KNIGHT)
    eq([], b.moves_from("e2"), "blocked pawn can't move or jump")
    b.place("d3", Color.BLACK, Kind.PAWN)
    b.place("f3", Color.WHITE, Kind.PAWN)
    eq(["d3"], b.moves_from("e2"), "captures only enemy diagonally")
    b.place("c5", Color.WHITE, Kind.PAWN)
    eq(["c6"], b.moves_from("c5"), "no double step off the start rank")
    b.place("h7", Color.BLACK, Kind.PAWN)
    b.place("h5", Color.WHITE, Kind.ROOK)
    eq(["h6"], b.moves_from("h7"), "double step needs both squares empty")


def test_move_turns_and_capture():
    b = ChessBoard.standard()
    eq(Color.WHITE, b.turn(), "white first")
    raises(WrongTurnError, lambda: b.move("e7", "e5"), "black can't move first")
    b.move("e2", "e4")
    eq(Color.BLACK, b.turn(), "turn flips")
    b.move("d7", "d5")
    b.move("e4", "d5")
    eq("wP", b.piece_at("d5"), "capture replaces the piece")
    eq(".", b.piece_at("e4"), "source emptied")
    raises(IllegalMoveError, lambda: b.move("d8", "d4"), "black queen d8 is stopped by the capture on d5, so d4 is illegal")
    raises(IllegalMoveError, lambda: b.move("e5", "e4"), "no piece")
    raises(ValueError, lambda: b.move("z9", "e4"), "invalid square")
