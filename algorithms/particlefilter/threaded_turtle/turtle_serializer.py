from turtle import Turtle

from threaded_turtle.thread_serializer import MethodCommand


class TurtleMethodCommand(MethodCommand):
    """The MethodCommand for Turtle commands

    Represents a method of turtle.Turtle.

    The class on which this is defined must have these two attributes:
     - turtle - instnce of Python built-in turtle.Turtle
     - thread_serializer - instance of ThreadSerializer

    """

    def _execute(self, threaded_turtle_instance, command_name, args, kwargs):
        threadless_turtle = threaded_turtle_instance.turtle

        method = getattr(threadless_turtle, command_name)
        return method(*args, **kwargs)

    def _get_thread_serializer(self, instance):
        return instance.thread_serializer


class TurtleCommander:
    def __init__(self, thread_serializer, turtle=None):
        self.thread_serializer = thread_serializer
        if turtle is None:
            self.turtle = Turtle()
        elif isinstance(turtle, TurtleCommander):
            self.turtle = turtle.turtle
        else:
            assert isinstance(turtle, Turtle)
            self.turtle = turtle

    forward = TurtleMethodCommand()
    fd = TurtleMethodCommand()
    backward = TurtleMethodCommand()
    bk = TurtleMethodCommand()
    back = TurtleMethodCommand()
    right = TurtleMethodCommand()
    rt = TurtleMethodCommand()
    left = TurtleMethodCommand()
    lt = TurtleMethodCommand()
    goto = TurtleMethodCommand()
    setpos = TurtleMethodCommand()
    setposition = TurtleMethodCommand()
    setx = TurtleMethodCommand()
    sety = TurtleMethodCommand()
    setheading = TurtleMethodCommand()
    seth = TurtleMethodCommand()
    home = TurtleMethodCommand()
    circle = TurtleMethodCommand()
    dot = TurtleMethodCommand()
    stamp = TurtleMethodCommand()
    clearstamp = TurtleMethodCommand()
    clearstamps = TurtleMethodCommand()
    undo = TurtleMethodCommand()
    speed = TurtleMethodCommand()
    position = TurtleMethodCommand()
    pos = TurtleMethodCommand()
    towards = TurtleMethodCommand()
    xcor = TurtleMethodCommand()
    ycor = TurtleMethodCommand()
    heading = TurtleMethodCommand()
    distance = TurtleMethodCommand()
    degrees = TurtleMethodCommand()
    radians = TurtleMethodCommand()
    pendown = TurtleMethodCommand()
    pd = TurtleMethodCommand()
    down = TurtleMethodCommand()
    penup = TurtleMethodCommand()
    pu = TurtleMethodCommand()
    up = TurtleMethodCommand()
    pensize = TurtleMethodCommand()
    width = TurtleMethodCommand()
    pen = TurtleMethodCommand()
    isdown = TurtleMethodCommand()
    color = TurtleMethodCommand()
    pencolor = TurtleMethodCommand()
    fillcolor = TurtleMethodCommand()
    filling = TurtleMethodCommand()
    begin_fill = TurtleMethodCommand()
    end_fill = TurtleMethodCommand()
    reset = TurtleMethodCommand()
    clear = TurtleMethodCommand()
    write = TurtleMethodCommand()
    showturtle = TurtleMethodCommand()
    st = TurtleMethodCommand()
    hideturtle = TurtleMethodCommand()
    ht = TurtleMethodCommand()
    isvisible = TurtleMethodCommand()
    shape = TurtleMethodCommand()
    resizemode = TurtleMethodCommand()
    shapesize = TurtleMethodCommand()
    turtlesize = TurtleMethodCommand()
    shearfactor = TurtleMethodCommand()
    settiltangle = TurtleMethodCommand()
    tiltangle = TurtleMethodCommand()
    tilt = TurtleMethodCommand()
    shapetransform = TurtleMethodCommand()
    get_shapepoly = TurtleMethodCommand()
    begin_poly = TurtleMethodCommand()
    end_poly = TurtleMethodCommand()
    get_poly = TurtleMethodCommand()
    clone = TurtleMethodCommand()
    getturtle = TurtleMethodCommand()
    getpen = TurtleMethodCommand()
    getscreen = TurtleMethodCommand()
    setundobuffer = TurtleMethodCommand()
    undobufferentries = TurtleMethodCommand()
