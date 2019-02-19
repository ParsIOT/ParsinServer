"""
This package brings the functionality of Python's "turtle" package
with threading.

Turtle does not support threading, so this implements a queue which
serializes the access to turtle API from the main thread, which is
easily accessible from the turtle threads.

Usage:

# Initialize one TurtleThreadSerializer
ctrl = ThreadSerializer()

# Initialize TurtleThread instances as you want
# each represents a separate turtle, running in a separate thread

def run(my_turtle, angle):
    my_turtle.left(angle)
    for i in range(10, 0, -1):
        my_turtle.left(5)
        my_turtle.forward(i)

turtle1 = TurtleThread(ctrl, target=run, args=(0,))
turtle2 = TurtleThread(ctrl, target=run, args=(90,))
turtle3 = TurtleThread(ctrl, target=run, args=(180,))
turtle4 = TurtleThread(ctrl, target=run, args=(270,))


# Start all turtles:
turtle1.start()
turtle2.start()
turtle3.start()
turtle4.start()

# Start the main controller:
ctrl.run_forever()


"""
from threaded_turtle.thread_serializer import ThreadSerializer
from threaded_turtle.turtle_thread import TurtleThread
