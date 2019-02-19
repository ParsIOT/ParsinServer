import threading

from threaded_turtle.turtle_serializer import TurtleCommander


class TurtleThread(threading.Thread):
    """A thread which knows how to handle a turtle ;)

    The constructor takes two additional arguments compared to vanilla Thread:
     - thread_serializer - the object which makes it possible to
                           run turtles in threads
     - turtle            - the optional turtle.TurtleCommander, if omitted,
                           a new one will be created

    There are two main ways to use it:
    - if the thread is implemented by overriding the run method, use self.turtle
      as if it was a normal turtle.Turtle which can run in a thread
    - if the thread is initialized with a "target" argument, the target will
      receive the turtle as the first argument, plus all args and kwargs

    """

    def __init__(self, thread_serializer, turtle=None, target=None, name=None,
                 args=(), kwargs={}, daemon=None):
        self.turtle = TurtleCommander(thread_serializer, turtle)
        if target is not None:
            args = (self.turtle,) + tuple(args)
        super().__init__(target=target, name=name, args=args, kwargs=kwargs, daemon=daemon)
