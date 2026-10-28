import java.util.*;

enum Player { X, O }

enum Status { IN_PROGRESS, X_WON, O_WON, DRAW }

final class TicTacToe {
    private final int n;
    private final Player[][] board;
    private final int[] rows, cols;
    private int diag, anti, moves;
    private Player turn = Player.X;
    private Status status = Status.IN_PROGRESS;

    TicTacToe(int n) {
        if (n < 3) throw new IllegalArgumentException("n must be >= 3");
        this.n = n;
        board = new Player[n][n];
        rows = new int[n];
        cols = new int[n];
    }

    Status move(Player p, int row, int col) {
        if (status != Status.IN_PROGRESS) throw new IllegalStateException("game over: " + status);
        if (row < 0 || row >= n || col < 0 || col >= n) throw new IllegalArgumentException("out of bounds");
        if (board[row][col] != null) throw new IllegalArgumentException("cell occupied");
        if (p != turn) throw new IllegalStateException("not " + p + "'s turn");

        board[row][col] = p;
        moves++;
        int d = p == Player.X ? 1 : -1;
        rows[row] += d;
        cols[col] += d;
        if (row == col) diag += d;
        if (row + col == n - 1) anti += d;

        int target = d * n;
        if (rows[row] == target || cols[col] == target || diag == target || anti == target) {
            status = p == Player.X ? Status.X_WON : Status.O_WON;
        } else if (moves == n * n) {
            status = Status.DRAW;
        }
        turn = turn == Player.X ? Player.O : Player.X;
        return status;
    }

    Status status() { return status; }

    char cell(int row, int col) {
        Player p = board[row][col];
        return p == null ? '.' : p.name().charAt(0);
    }
}
