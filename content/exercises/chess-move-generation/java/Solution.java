import java.util.*;

enum Color { WHITE, BLACK }

enum Kind { KING, QUEEN, ROOK, BISHOP, KNIGHT, PAWN }

// TODO: a Square value type, an abstract Piece with one subclass per kind (or a data-driven design)

final class ChessBoard {
    static ChessBoard empty() { throw new UnsupportedOperationException("TODO"); }

    static ChessBoard standard() { throw new UnsupportedOperationException("TODO"); }

    void place(String square, Color color, Kind kind) { throw new UnsupportedOperationException("TODO"); }

    String pieceAt(String square) { throw new UnsupportedOperationException("TODO"); }

    List<String> movesFrom(String square) { throw new UnsupportedOperationException("TODO"); }

    void move(String from, String to) { throw new UnsupportedOperationException("TODO"); }

    Color turn() { throw new UnsupportedOperationException("TODO"); }
}
