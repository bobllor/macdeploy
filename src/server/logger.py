from logging import Formatter, Logger, StreamHandler, FileHandler
from logging import Handler
from logging import DEBUG
from datetime import datetime
from typing import TypedDict, Literal, TextIO
from configuration import SERVER_LOGS_PATH
from pathlib import Path
import sys

# this value will be the same as soon as the server is launched.
# this keeps all logs in one day on the same day.
DEFAULT_FILENAME: datetime = datetime.now().strftime("server-%Y-%m-%d.log")
DEFAULT_DATEFMT: Literal["%Y-%m-%d %H:%M:%S"] = "%Y-%m-%d %H:%M:%S"

class LogLevelOptions(TypedDict):
    log_level: str | int
    stream_level: str | int
    file_level: str | int

class Log(Logger):
    def __init__(self, 
    name: str = __name__,
    *, 
    log_path: Path | str = None,
    stream: TextIO = sys.stdout,
    levels: LogLevelOptions = {},
    logfmt: str = "%(asctime)s:%(filename)s:%(name)s [%(levelname)s] %(message)s",
    datefmt: str = DEFAULT_DATEFMT):
        '''Create a new instance of the Log class. It extends the `Logger` class.
        
        Parameters
        ----------
            name: str default `__name__`
                The name of the logger. By default, it uses the __name__ variable, or
                in other words __main__.
            
            log_path: Path | str, default `None`
                The log directory path. By default it is None, meaning it is not written to any directory.

            stream: TextIO default `sys.stdout`
                The stream output of the StreamHandler. By default it prints out to the console, sys.stdout.
            
            levels: LogLevelOptions default `{}`
                A dictionary that holds the logging levels of the logger, stream handler, and file handler.
                The keys are "log_level", "stream_level", and "file_level". By default **all levels are debug**.
            
            logfmt: str default `TIME FILENAME LOGNAME [LEVEL] MESSAGE`
                The format of the log message.
            
            datefmt: str default `YY-MM-DD HH-MM-SS`
                The format for the date.
        '''
        super().__init__(name)

        stream_handler: StreamHandler = StreamHandler(stream)
        file_handler: FileHandler = None
        formatter: Formatter = Formatter(fmt=logfmt, datefmt=datefmt)

        if log_path is not None:
            new_log_path: Path = Path("")
            if isinstance(log_path, str):
                new_log_path = Path(log_path)
            elif isinstance(log_path, Path):
                new_log_path = log_path
            
            if not new_log_path.exists():
                new_log_path.mkdir(parents=True, exist_ok=True)

            log_file: Path = new_log_path / DEFAULT_FILENAME
            file_handler = FileHandler(log_file)

        handlers: list[tuple[str, Handler]] = [
            ("stream_level", stream_handler)
        ]
        if file_handler is not None:
            handlers.append(("file_level", file_handler))

        for level_key, hdlr in handlers:
            hdlr.setFormatter(formatter)
            hdlr.setLevel(levels.get(level_key, DEBUG))

            self.addHandler(hdlr)

        self.setLevel(levels.get("log_level", DEBUG))