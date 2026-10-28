import java.util.*;

public class Tests {
    static void testStandardOpeningMoves() {
        ChessBoard b = ChessBoard.standard();
        T.eq(List.of("a3", "c3"), b.movesFrom("b1"), "knight jumps over pawns");
        T.eq(List.of("e3", "e4"), b.movesFrom("e2"), "pawn single + double step");
        T.eq(List.of(), b.movesFrom("a1"), "rook boxed in");
        T.eq(List.of("f6", "h6"), b.movesFrom("g8"), "black knight");
        T.eq(List.of("d5", "d6"), b.movesFrom("d7"), "black pawn moves down");
        T.eq(List.of(), b.movesFrom("e5"), "empty square");
        T.eq("wK", b.pieceAt("e1"), "white king");
        T.eq("bN", b.pieceAt("b8"), "black knight letter is N");
    }

    static void testSlidersOnEmptyBoard() {
        ChessBoard b = ChessBoard.empty();
        b.place("d4", Color.WHITE, Kind.ROOK);
        T.eq(14, b.movesFrom("d4").size(), "rook in the middle");
        b.place("a1", Color.WHITE, Kind.BISHOP);
        T.eq(List.of("b2", "c3"), b.movesFrom("a1"), "bishop stops before its own rook on d4");
        b.place("h8", Color.BLACK, Kind.QUEEN);
        T.eq(List.of("a8", "b8", "c8", "d8", "e8", "f8", "g8", "h1", "h2", "h3", "h4", "h5", "h6", "h7"),
                filter(b.movesFrom("h8"), "8", "h"), "queen rank + file");
        T.check(b.movesFrom("h8").contains("d4"), "queen diagonal reaches d4 (capture)");
        T.check(!b.movesFrom("h8").contains("c3"), "queen stops at the captured piece");
    }

    static List<String> filter(List<String> sqs, String rank, String file) {
        List<String> out = new ArrayList<>();
        for (String s : sqs) if (s.endsWith(rank) || s.startsWith(file)) out.add(s);
        return out;
    }

    static void testBlockingAndCaptures() {
        ChessBoard b = ChessBoard.empty();
        b.place("a1", Color.WHITE, Kind.ROOK);
        b.place("a4", Color.BLACK, Kind.PAWN);
        b.place("c1", Color.WHITE, Kind.KNIGHT);
        T.eq(List.of("a2", "a3", "a4", "b1"), b.movesFrom("a1"), "capture a4, blocked by own knight on c1");
    }

    static void testKingAndKnight() {
        ChessBoard b = ChessBoard.empty();
        b.place("e4", Color.WHITE, Kind.KING);
        T.eq(List.of("d3", "d4", "d5", "e3", "e5", "f3", "f4", "f5"), b.movesFrom("e4"), "king in the centre");
        b.place("a1", Color.BLACK, Kind.KING);
        T.eq(List.of("a2", "b1", "b2"), b.movesFrom("a1"), "king in the corner");
        b.place("g1", Color.WHITE, Kind.KNIGHT);
        T.eq(List.of("e2", "f3", "h3"), b.movesFrom("g1"), "knight near the edge");
    }

    static void testPawnRules() {
        ChessBoard b = ChessBoard.empty();
        b.place("e2", Color.WHITE, Kind.PAWN);
        b.place("e3", Color.BLACK, Kind.KNIGHT);
        T.eq(List.of(), b.movesFrom("e2"), "blocked pawn can't move or jump");
        b.place("d3", Color.BLACK, Kind.PAWN);
        b.place("f3", Color.WHITE, Kind.PAWN);
        T.eq(List.of("d3"), b.movesFrom("e2"), "captures only enemy diagonally");
        b.place("c5", Color.WHITE, Kind.PAWN);
        T.eq(List.of("c6"), b.movesFrom("c5"), "no double step off the start rank");
        b.place("h7", Color.BLACK, Kind.PAWN);
        b.place("h5", Color.WHITE, Kind.ROOK);
        T.eq(List.of("h6"), b.movesFrom("h7"), "double step needs both squares empty");
    }

    static void testMoveTurnsAndCapture() {
        ChessBoard b = ChessBoard.standard();
        T.eq(Color.WHITE, b.turn(), "white first");
        T.throwsEx(IllegalStateException.class, () -> b.move("e7", "e5"), "black can't move first");
        b.move("e2", "e4");
        T.eq(Color.BLACK, b.turn(), "turn flips");
        b.move("d7", "d5");
        b.move("e4", "d5");
        T.eq("wP", b.pieceAt("d5"), "capture replaces the piece");
        T.eq(".", b.pieceAt("e4"), "source emptied");
        T.throwsEx(IllegalArgumentException.class, () -> b.move("d8", "d4"), "black queen d8 is stopped by the capture on d5, so d4 is illegal");
        T.throwsEx(IllegalArgumentException.class, () -> b.move("e5", "e4"), "no piece");
        T.throwsEx(IllegalArgumentException.class, () -> b.move("z9", "e4"), "invalid square");
    }
}
