import java.util.*;

interface Dice {
    int roll();
}

record TurnResult(String player, int roll, int from, int to, boolean won) {}

final class Board {
    private final int size;
    private final Map<Integer, Integer> jumps = new HashMap<>();

    Board(int size, Map<Integer, Integer> snakes, Map<Integer, Integer> ladders) {
        if (size < 2) throw new IllegalArgumentException("size too small");
        this.size = size;
        snakes.forEach((head, tail) -> add(head, tail, head > tail, "snake"));
        ladders.forEach((bottom, top) -> add(bottom, top, bottom < top, "ladder"));
        for (int end : jumps.values())
            if (jumps.containsKey(end)) throw new IllegalArgumentException("jump ends on another jump's start: " + end);
    }

    private void add(int from, int to, boolean rightDirection, String kind) {
        if (!rightDirection) throw new IllegalArgumentException(kind + " goes the wrong way: " + from + "->" + to);
        if (from < 1 || from >= size || to < 1 || to > size) throw new IllegalArgumentException(kind + " out of range");
        if (jumps.putIfAbsent(from, to) != null) throw new IllegalArgumentException("two jumps start at " + from);
    }

    int size() { return size; }

    /** Where a player who lands on `cell` ends up. */
    int resolve(int cell) { return jumps.getOrDefault(cell, cell); }
}

final class Game {
    private final Board board;
    private final List<String> players;
    private final Dice dice;
    private final Map<String, Integer> pos = new HashMap<>();
    private int current;
    private int sixes;
    private int turnStart;
    private String winner;

    Game(Board board, List<String> players, Dice dice) {
        if (players.size() < 2) throw new IllegalArgumentException("need at least 2 players");
        this.board = board;
        this.players = List.copyOf(players);
        this.dice = dice;
        for (String p : players) pos.put(p, 0);
        turnStart = 0;
    }

    TurnResult playTurn() {
        if (winner != null) throw new IllegalStateException("game over");
        String p = players.get(current);
        int from = pos.get(p);
        if (sixes == 0) turnStart = from;
        int roll = dice.roll();

        if (roll == 6 && ++sixes == 3) {                 // penalty: undo the whole turn
            pos.put(p, turnStart);
            passTurn();
            return new TurnResult(p, roll, from, turnStart, false);
        }

        int target = from + roll;
        int to = target > board.size() ? from : board.resolve(target);
        pos.put(p, to);

        if (to == board.size()) {
            winner = p;
            return new TurnResult(p, roll, from, to, true);
        }
        if (roll != 6) passTurn();
        return new TurnResult(p, roll, from, to, false);
    }

    private void passTurn() {
        current = (current + 1) % players.size();
        sixes = 0;
    }

    int position(String player) { return pos.get(player); }

    Optional<String> winner() { return Optional.ofNullable(winner); }
}
