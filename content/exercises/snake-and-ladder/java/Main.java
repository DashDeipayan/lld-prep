import java.util.*;

public class Main {
    public static void main(String[] args) {
        Board board = new Board(100, Map.of(99, 7, 62, 19, 48, 26), Map.of(2, 38, 28, 84, 51, 67, 71, 91));
        Random rnd = new Random(42);
        Game g = new Game(board, List.of("alice", "bob"), () -> 1 + rnd.nextInt(6));
        int turns = 0;
        while (g.winner().isEmpty() && turns < 1000) {
            TurnResult r = g.playTurn();
            if (turns++ < 8) System.out.println(r);
        }
        System.out.println("... winner = " + g.winner().orElse("none") + " after " + turns + " rolls");
    }
}
