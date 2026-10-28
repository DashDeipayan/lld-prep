import java.util.*;

public class Main {
    public static void main(String[] args) throws Exception {
        BoundedBlockingQueue<Integer> q = new BoundedBlockingQueue<>(2);
        Thread consumer = new Thread(() -> {
            try {
                Optional<Integer> x;
                while ((x = q.take()).isPresent()) System.out.println("consumed " + x.get());
                System.out.println("consumer: queue closed and drained");
            } catch (InterruptedException ignored) {}
        });
        consumer.start();
        for (int i = 1; i <= 5; i++) {
            q.put(i);
            System.out.println("produced " + i);
        }
        q.close();
        consumer.join();
    }
}
