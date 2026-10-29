import threading

from solution import BoundedBlockingQueue

q = BoundedBlockingQueue(2)


def consumer():
    while True:
        x = q.take()
        if x is None:
            break
        print("consumed", x)
    print("consumer: queue closed and drained")


t = threading.Thread(target=consumer)
t.start()
for i in range(1, 6):
    q.put(i)
    print("produced", i)
q.close()
t.join()
