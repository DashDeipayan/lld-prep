import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

record Message(long offset, String payload) {}

final class Broker {
    /** One lock per topic: different topics never contend. */
    private static final class Topic {
        final List<String> log = new ArrayList<>();
        final Map<String, Long> offsets = new HashMap<>();
    }

    private final ConcurrentHashMap<String, Topic> topics = new ConcurrentHashMap<>();

    void createTopic(String topic) {
        if (topics.putIfAbsent(topic, new Topic()) != null) throw new IllegalStateException("topic exists: " + topic);
    }

    long publish(String topic, String payload) {
        Topic t = topic(topic);
        synchronized (t) {
            t.log.add(payload);
            return t.log.size() - 1;
        }
    }

    void subscribe(String topic, String group, boolean fromBeginning) {
        Topic t = topic(topic);
        synchronized (t) {
            t.offsets.putIfAbsent(group, fromBeginning ? 0L : t.log.size());
        }
    }

    List<Message> poll(String topic, String group, int max) {
        Topic t = topic(topic);
        synchronized (t) {
            long from = offset(t, group);
            long to = Math.min(t.log.size(), from + Math.max(0, max));
            List<Message> out = new ArrayList<>((int) (to - from));
            for (long i = from; i < to; i++) out.add(new Message(i, t.log.get((int) i)));
            t.offsets.put(group, to);
            return out;
        }
    }

    void seek(String topic, String group, long offset) {
        Topic t = topic(topic);
        synchronized (t) {
            offset(t, group);
            if (offset < 0 || offset > t.log.size()) throw new IllegalArgumentException("offset out of range: " + offset);
            t.offsets.put(group, offset);
        }
    }

    long lag(String topic, String group) {
        Topic t = topic(topic);
        synchronized (t) {
            return t.log.size() - offset(t, group);
        }
    }

    private Topic topic(String name) {
        Topic t = topics.get(name);
        if (t == null) throw new NoSuchElementException("unknown topic " + name);
        return t;
    }

    private static long offset(Topic t, String group) {
        Long o = t.offsets.get(group);
        if (o == null) throw new IllegalStateException("group not subscribed: " + group);
        return o;
    }
}
