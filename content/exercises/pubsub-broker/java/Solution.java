import java.util.*;

record Message(long offset, String payload) {}

final class Broker {
    void createTopic(String topic) { throw new UnsupportedOperationException("TODO"); }

    long publish(String topic, String payload) { throw new UnsupportedOperationException("TODO"); }

    void subscribe(String topic, String group, boolean fromBeginning) { throw new UnsupportedOperationException("TODO"); }

    List<Message> poll(String topic, String group, int max) { throw new UnsupportedOperationException("TODO"); }

    void seek(String topic, String group, long offset) { throw new UnsupportedOperationException("TODO"); }

    long lag(String topic, String group) { throw new UnsupportedOperationException("TODO"); }
}
