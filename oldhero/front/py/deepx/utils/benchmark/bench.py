import time

class Bench:
    def __init__(self, name: str):
        self.name = name
    
    def __call__(self, func):
        def wrapper(*args, **kwargs):   
            result = func(*args, **kwargs)
            print(f"{self.name} took {time.time() - start_time} seconds to run")
            return result