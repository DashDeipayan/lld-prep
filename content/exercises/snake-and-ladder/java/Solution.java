import java.util.*;

interface Dice {
    int roll();
}

record TurnResult(String player, int roll, int from, int to, boolean won) {}

final class Board {
    Board(int size, Map<Integer, Integer> snakes, Map<Integer, Integer> ladders) {
        // TODO: validate and store
    }
}

final class Game {
    Game(Board board, List<String> players, Dice dice) {
        // TODO
    }

    TurnResult playTurn() { throw new UnsupportedOperationException("TODO"); }

    int position(String player) { throw new UnsupportedOperationException("TODO"); }

    Optional<String> winner() { throw new UnsupportedOperationException("TODO"); }
}
