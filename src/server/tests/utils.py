from logger import Log

def get_log(path: str) -> Log:
    return Log(log_path=path)