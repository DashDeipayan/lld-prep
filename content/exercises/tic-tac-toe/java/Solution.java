import java.util.*;

enum Player { X, O }

enum Status { IN_PROGRESS, X_WON, O_WON, DRAW }

final class TicTacToe {
    TicTacToe(int n) { /* TODO */ }

    Status move(Player p, int row, int col) { throw new UnsupportedOperationException("TODO"); }

    Status status() { throw new UnsupportedOperationException("TODO"); }

    char cell(int row, int col) { throw new UnsupportedOperationException("TODO"); }
}
