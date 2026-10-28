import java.util.*;

public class Tests {
    static Dice script(int... rolls) {
        int[] i = {0};
        return () -> rolls[i[0]++];
    }

    static Board board() { return new Board(20, Map.of(17, 5), Map.of(3, 11)); }

    static void testBoardValidation() {
        T.throwsEx(IllegalArgumentException.class, () -> new Board(20, Map.of(5, 9), Map.of()), "snake going up");
        T.throwsEx(IllegalArgumentException.class, () -> new Board(20, Map.of(), Map.of(9, 5)), "ladder going down");
        T.throwsEx(IllegalArgumentException.class, () -> new Board(20, Map.of(25, 3), Map.of()), "out of range");
        T.throwsEx(IllegalArgumentException.class, () -> new Board(20, Map.of(20, 3), Map.of()), "jump from final cell");
        T.throwsEx(IllegalArgumentException.class, () -> new Board(20, Map.of(9, 2), Map.of(9, 15)), "two jumps from same cell");
        T.throwsEx(IllegalArgumentException.class, () -> new Board(20, Map.of(15, 4), Map.of(4, 12)), "chain: ladder 4->12, snake 15->4 ends on 4");
        T.throwsEx(IllegalArgumentException.class, () -> new Game(board(), List.of("solo"), script(1)), "need 2 players");
    }

    static void testBasicTurnsAlternate() {
        Game g = new Game(board(), List.of("A", "B"), script(2, 4, 1));
        T.eq(new TurnResult("A", 2, 0, 2, false), g.playTurn(), "A");
        T.eq(new TurnResult("B", 4, 0, 4, false), g.playTurn(), "B");
        T.eq(new TurnResult("A", 1, 2, 11, false), g.playTurn(), "A lands on ladder 3 -> 11");
        T.eq(11, g.position("A"), "position");
    }

    static void testSnake() {
        Game g = new Game(board(), List.of("A", "B"), script(5, 1, 6, 6));
        g.playTurn();
        g.playTurn();
        T.eq(new TurnResult("A", 6, 5, 11, false), g.playTurn(), "A: 5 -> 11 (six, rolls again)");
        T.eq(new TurnResult("A", 6, 11, 5, false), g.playTurn(), "A: 11 -> 17, snake down to 5");
    }

    static void testSixGivesExtraRoll() {
        Game g = new Game(board(), List.of("A", "B"), script(6, 1, 2));
        g.playTurn();
        T.eq("A", g.playTurn().player(), "A rolls again after a six");
        T.eq("B", g.playTurn().player(), "then B");
    }

    static void testThreeSixesPenalty() {
        Game g = new Game(new Board(40, Map.of(), Map.of()), List.of("A", "B"), script(2, 1, 6, 6, 6, 3));
        g.playTurn();
        g.playTurn();
        g.playTurn();
        g.playTurn();
        TurnResult third = g.playTurn();
        T.eq(new TurnResult("A", 6, 14, 2, false), third, "third six sends A back to the start of the turn");
        T.eq(2, g.position("A"), "A back at 2");
        T.eq("B", g.playTurn().player(), "turn passes after the penalty");
    }

    static void testExactFinishAndWinner() {
        Game g = new Game(new Board(10, Map.of(), Map.of()), List.of("A", "B"), script(5, 1, 6, 1, 4, 4));
        g.playTurn();
        g.playTurn();
        T.eq(new TurnResult("A", 6, 5, 5, false), g.playTurn(), "overshoot: 5+6 > 10, stay (six still grants a reroll)");
        T.eq(new TurnResult("A", 1, 5, 6, false), g.playTurn(), "A again");
        g.playTurn();
        TurnResult win = g.playTurn();
        T.eq(new TurnResult("A", 4, 6, 10, true), win, "A lands exactly on 10");
        T.eq(Optional.of("A"), g.winner(), "winner");
        T.throwsEx(IllegalStateException.class, g::playTurn, "no turns after game over");
    }
}
