import multiprocessing

preload_app = True # i got this from a stack overflow in regards to thread locks.
loglevel = "info"
workers = (multiprocessing.cpu_count() * 2) + 1