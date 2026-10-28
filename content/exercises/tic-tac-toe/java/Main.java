public class Main {
    public static void main(String[] args) {
        TicTacToe g = new TicTacToe(3);
        int[][] moves = {{0, 0}, {1, 1}, {0, 1}, {2, 2}, {0, 2}};
        Player p = Player.X;
        for (int[] m : moves) {
            System.out.println(p + " -> (" + m[0] + "," + m[1] + "): " + g.move(p, m[0], m[1]));
            p = p == Player.X ? Player.O : Player.X;
        }
        for (int r = 0; r < 3; r++) {
            StringBuilder sb = new StringBuilder();
            for (int c = 0; c < 3; c++) sb.append(g.cell(r, c));
            System.out.println(sb);
        }
    }
}
