"""
Serialize thread execution using a queue.

When multiple thread are needed to be used in an environment where
multithreading is not supported, this module provides the functionality
for executing threads which send their commands to a queue which is then
executed sequentially in the main thread.

The functionality consists of:
- thread serializer - which puts commands in a queue when execute_command
                      is called from different threads, and then executes
                      the commands after reading them from the queue in the
                      main thread
- method command    - descriptor which creates a method to be called from a
                      thread; when called, the method creates a command in
                      thread serializer and waits for it to be executed, so
                      it seems as if it was executed within the thread
"""
import queue
import threading


class ThreadSerializer:
    """Sequentially executes commands it finds in its queue"""

    def __init__(self):
        self._queue = queue.Queue()

    def run_forever(self, queue_timeout=None):
        """Read commands from the queue and execute them forever

        If queue_timeout is given, this method will raise
        queue.Empty if there is noting in the queue for that number
        if seconds.
        """
        while True:
            command = self._queue.get(timeout=queue_timeout)
            command.execute()

    def execute_command(self, command):
        """Put command in the queue, wait for the result and return it"""
        self._queue.put(command)
        return command.wait_for_result()


class MethodCommand:
    """A descriptor which turns methods into commands

    Abstract class. Implement methods _execute and _get_thread_serializer:
    * _execute executes a command. It will always be executed from the main thread.
    * _get_thread_serializer return the thread serialized which executes commands

    Usage example:

        If a class defines a MethodCommand named foo, like this:

        class ClassTwo:
            foo = ImplementationOfMethodCommand()

        then foo can be called as a method and it will execute
        a command named "foo" with arguments passed to the method
    """

    def __init__(self, command_name=None):
        """
        Arguments:
            command_name: name of the command, defaults to the descriptor name

        """
        self._command_name = command_name

    def _execute(self, instance, command_name, args, kwargs):
        """Override this method. It should return a Command instance."""
        raise NotImplementedError()

    def _get_thread_serializer(self, instance):
        """Override this method. It should return a ThreadSerializer instance."""
        raise NotImplementedError()

    def _get_redirected_func(self, instance):
        serializer = self._get_thread_serializer(instance)

        def func(*a, **kw):
            execute = lambda: self._execute(instance, self._command_name, a, kw)
            return serializer.execute_command(_Command(execute))

        func.__qualname__ = "command<{}>".format(self._command_name)
        return func

    def __get__(self, instance, owner):
        if instance is not None:
            return self._get_redirected_func(instance)
        else:
            def unbound_f(inst, *a, **kw):
                bound_f = self._get_redirected_func(inst)
                return bound_f(*a, **kw)

            unbound_f.__qualname__ = "unbound command<{}>".format(self._command_name)
            return unbound_f

    def __set_name__(self, owner, name):
        if self._command_name is None:
            self._command_name = name


class _Command:
    """A command to be executed in another thread"""

    def __init__(self, execute_callback):
        self._result_ready_event = threading.Event()
        self._result = None
        self._execute_callback = execute_callback

    def execute(self):
        self._result = self._execute_callback()
        self._result_ready_event.set()

    def wait_for_result(self):
        self._result_ready_event.wait()
        return self._result
