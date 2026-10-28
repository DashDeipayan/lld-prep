public class Tests {
    static Status play(TicTacToe g, int[][] moves) {
        Player p = Player.X;
        Status s = Status.IN_PROGRESS;
        for (int[] m : moves) { s = g.move(p, m[0], m[1]); p = p == Player.X ? Player.O : Player.X; }
        return s;
    }

    static void testRowWin() {
        TicTacToe g = new TicTacToe(3);
        T.eq(Status.X_WON, play(g, new int[][] {{0, 0}, {1, 0}, {0, 1}, {1, 1}, {0, 2}}), "X wins top row");
        T.eq('X', g.cell(0, 2), "cell");
        T.eq('.', g.cell(2, 2), "empty cell");
    }

    static void testColumnAndDiagonalWins() {
        T.eq(Status.O_WON, play(new TicTacToe(3), new int[][] {{0, 0}, {0, 1}, {1, 0}, {1, 1}, {2, 2}, {2, 1}}), "O wins middle column");
        T.eq(Status.X_WON, play(new TicTacToe(3), new int[][] {{0, 0}, {0, 1}, {1, 1}, {0, 2}, {2, 2}}), "X main diagonal");
        T.eq(Status.X_WON, play(new TicTacToe(3), new int[][] {{0, 2}, {0, 0}, {1, 1}, {0, 1}, {2, 0}}), "X anti-diagonal");
    }

    static void testLargerBoard() {
        TicTacToe g = new TicTacToe(4);
        Status s = play(g, new int[][] {{0, 3}, {0, 0}, {1, 2}, {1, 0}, {2, 1}, {2, 0}, {3, 0}});
        T.eq(Status.X_WON, s, "4x4 anti-diagonal completed by X at (3,0)");
    }

    static void testDraw() {
        TicTacToe g = new TicTacToe(3);
        Status s = play(g, new int[][] {{0, 0}, {0, 1}, {0, 2}, {1, 1}, {1, 0}, {1, 2}, {2, 1}, {2, 0}, {2, 2}});
        T.eq(Status.DRAW, s, "full board, no line");
    }

    static void testInvalidMovesChangeNothing() {
        TicTacToe g = new TicTacToe(3);
        T.throwsEx(IllegalStateException.class, () -> g.move(Player.O, 0, 0), "X moves first");
        g.move(Player.X, 1, 1);
        T.throwsEx(IllegalArgumentException.class, () -> g.move(Player.O, 1, 1), "occupied");
        T.throwsEx(IllegalArgumentException.class, () -> g.move(Player.O, 3, 0), "out of bounds");
        T.throwsEx(IllegalStateException.class, () -> g.move(Player.X, 0, 0), "X can't move twice");
        T.eq(Status.IN_PROGRESS, g.move(Player.O, 0, 0), "O's valid move still accepted");
        T.throwsEx(IllegalArgumentException.class, () -> new TicTacToe(2), "n >= 3");
    }

    static void testNoMovesAfterGameOver() {
        TicTacToe g = new TicTacToe(3);
        play(g, new int[][] {{0, 0}, {1, 0}, {0, 1}, {1, 1}, {0, 2}});
        T.throwsEx(IllegalStateException.class, () -> g.move(Player.O, 2, 2), "game over");
        T.eq(Status.X_WON, g.status(), "status sticks");
    }
}
