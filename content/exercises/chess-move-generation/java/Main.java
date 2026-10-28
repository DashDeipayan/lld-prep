public class Main {
    public static void main(String[] args) {
        ChessBoard b = ChessBoard.standard();
        System.out.println("b1 knight: " + b.movesFrom("b1"));
        System.out.println("e2 pawn:   " + b.movesFrom("e2"));
        b.move("e2", "e4");
        b.move("d7", "d5");
        System.out.println("e4 pawn can capture: " + b.movesFrom("e4"));
        b.move("e4", "d5");
        System.out.println("d5 now holds " + b.pieceAt("d5") + ", turn = " + b.turn());
        System.out.println("d1 queen:  " + b.movesFrom("d1"));
    }
}
