from logger import Log, LogLevelOptions
from logging import DEBUG, CRITICAL

LOG_OPTIONS: LogLevelOptions = {
    "log_level": CRITICAL
}

def get_log(path: str, *, levels: LogLevelOptions = None) -> Log:
    if not levels:
        levels = {"log_level": DEBUG}

    return Log(log_path=path, levels=levels)