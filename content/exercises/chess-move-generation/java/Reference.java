import java.util.*;

enum Color { WHITE, BLACK }

enum Kind { KING, QUEEN, ROOK, BISHOP, KNIGHT, PAWN }

record Square(int file, int rank) {
    static Square parse(String s) {
        if (s == null || s.length() != 2) throw new IllegalArgumentException("invalid square " + s);
        int f = s.charAt(0) - 'a', r = s.charAt(1) - '1';
        if (f < 0 || f > 7 || r < 0 || r > 7) throw new IllegalArgumentException("invalid square " + s);
        return new Square(f, r);
    }

    /** null when off the board */
    Square offset(int df, int dr) {
        int f = file + df, r = rank + dr;
        return f < 0 || f > 7 || r < 0 || r > 7 ? null : new Square(f, r);
    }

    @Override public String toString() { return "" + (char) ('a' + file) + (char) ('1' + rank); }
}

abstract class Piece {
    static final int[][] STRAIGHT = {{1, 0}, {-1, 0}, {0, 1}, {0, -1}};
    static final int[][] DIAGONAL = {{1, 1}, {1, -1}, {-1, 1}, {-1, -1}};
    static final int[][] ALL = {{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {1, -1}, {-1, 1}, {-1, -1}};
    static final int[][] KNIGHT = {{1, 2}, {2, 1}, {2, -1}, {1, -2}, {-1, -2}, {-2, -1}, {-2, 1}, {-1, 2}};

    final Color color;
    Piece(Color color) { this.color = color; }

    abstract Kind kind();
    abstract List<Square> candidateMoves(ChessBoard b, Square from);

    /** Walk each direction up to `range` squares; stop at the first piece, including it if it's an enemy. */
    protected List<Square> walk(ChessBoard b, Square from, int[][] dirs, int range) {
        List<Square> out = new ArrayList<>();
        for (int[] d : dirs) {
            Square s = from;
            for (int i = 0; i < range; i++) {
                s = s.offset(d[0], d[1]);
                if (s == null) break;
                Piece p = b.at(s);
                if (p == null) { out.add(s); continue; }
                if (p.color != color) out.add(s);
                break;
            }
        }
        return out;
    }

    static Piece of(Color c, Kind k) {
        return switch (k) {
            case KING -> new Stepper(c, k, ALL, 1);
            case QUEEN -> new Stepper(c, k, ALL, 7);
            case ROOK -> new Stepper(c, k, STRAIGHT, 7);
            case BISHOP -> new Stepper(c, k, DIAGONAL, 7);
            case KNIGHT -> new Stepper(c, k, KNIGHT, 1);
            case PAWN -> new Pawn(c);
        };
    }
}

/** King, queen, rook, bishop and knight differ only in directions and range. */
final class Stepper extends Piece {
    private final Kind kind;
    private final int[][] dirs;
    private final int range;
    Stepper(Color c, Kind kind, int[][] dirs, int range) { super(c); this.kind = kind; this.dirs = dirs; this.range = range; }
    Kind kind() { return kind; }
    List<Square> candidateMoves(ChessBoard b, Square from) { return walk(b, from, dirs, range); }
}

final class Pawn extends Piece {
    Pawn(Color c) { super(c); }
    Kind kind() { return Kind.PAWN; }

    List<Square> candidateMoves(ChessBoard b, Square from) {
        List<Square> out = new ArrayList<>();
        int dir = color == Color.WHITE ? 1 : -1, startRank = color == Color.WHITE ? 1 : 6;
        Square one = from.offset(0, dir);
        if (one != null && b.at(one) == null) {
            out.add(one);
            Square two = one.offset(0, dir);
            if (from.rank() == startRank && two != null && b.at(two) == null) out.add(two);
        }
        for (int df : new int[] {-1, 1}) {
            Square c = from.offset(df, dir);
            if (c != null && b.at(c) != null && b.at(c).color != color) out.add(c);
        }
        return out;
    }
}

final class ChessBoard {
    private final Piece[][] grid = new Piece[8][8];
    private Color turn = Color.WHITE;

    static ChessBoard empty() { return new ChessBoard(); }

    static ChessBoard standard() {
        ChessBoard b = new ChessBoard();
        Kind[] back = {Kind.ROOK, Kind.KNIGHT, Kind.BISHOP, Kind.QUEEN, Kind.KING, Kind.BISHOP, Kind.KNIGHT, Kind.ROOK};
        for (int f = 0; f < 8; f++) {
            b.grid[f][0] = Piece.of(Color.WHITE, back[f]);
            b.grid[f][1] = Piece.of(Color.WHITE, Kind.PAWN);
            b.grid[f][6] = Piece.of(Color.BLACK, Kind.PAWN);
            b.grid[f][7] = Piece.of(Color.BLACK, back[f]);
        }
        return b;
    }

    Piece at(Square s) { return grid[s.file()][s.rank()]; }

    void place(String square, Color color, Kind kind) {
        Square s = Square.parse(square);
        grid[s.file()][s.rank()] = Piece.of(color, kind);
    }

    String pieceAt(String square) {
        Piece p = at(Square.parse(square));
        if (p == null) return ".";
        String letter = p.kind() == Kind.KNIGHT ? "N" : p.kind().name().substring(0, 1);
        return (p.color == Color.WHITE ? "w" : "b") + letter;
    }

    List<String> movesFrom(String square) {
        Square s = Square.parse(square);
        Piece p = at(s);
        if (p == null) return List.of();
        List<String> out = new ArrayList<>();
        for (Square t : p.candidateMoves(this, s)) out.add(t.toString());
        Collections.sort(out);
        return out;
    }

    void move(String from, String to) {
        Square f = Square.parse(from), t = Square.parse(to);
        Piece p = at(f);
        if (p == null) throw new IllegalArgumentException("no piece at " + from);
        if (p.color != turn) throw new IllegalStateException("it is " + turn + "'s turn");
        if (!p.candidateMoves(this, f).contains(t)) throw new IllegalArgumentException("illegal move " + from + "-" + to);
        grid[t.file()][t.rank()] = p;
        grid[f.file()][f.rank()] = null;
        turn = turn == Color.WHITE ? Color.BLACK : Color.WHITE;
    }

    Color turn() { return turn; }
}
