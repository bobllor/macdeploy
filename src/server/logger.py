from logging import Formatter, Logger, StreamHandler, FileHandler, setLoggerClass
from logging import Handler
from logging import DEBUG
from datetime import datetime
from typing import TypedDict, Literal
from system.vars import Vars
from pathlib import Path

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
    log_path: Path | str = Vars.SERVER_LOG_PATH.value,
    levels: LogLevelOptions = {},
    logfmt: str = "%(asctime)s:%(filename)s:%(name)s [%(levelname)s] %(message)s",
    datefmt: str = DEFAULT_DATEFMT):
        '''Create a new instance of the Log class. It extends the `Logger` class.
        
        Parameters
        ----------
            name: str default `__name__`
                The name of the logger. By default, it uses the __name__ variable, or
                in other words __main__.
            
            log_path: Path | str, default `./logs/server`
                The log file name. By default it names the logs based on the current date
                and is stored in the logs folder, for example: `logs/server-2025-05-05/server-2000-11-01.log`.
            
            levels: LogLevelOptions default `{}`
                A dictionary that holds the logging levels of the logger, stream handler, and file handler.
                The keys are "log_level", "stream_level", and "file_level". By default **all levels are debug**.
            
            logfmt: str default `TIME FILENAME LOGNAME [LEVEL] MESSAGE`
                The format of the log message.
            
            datefmt: str default `YY-MM-DD HH-MM-SS`
                The format for the date.
        '''
        super().__init__(name)

        new_log_path: Path = Path("")
        if isinstance(log_path, str):
            new_log_path = Path(log_path)
        elif isinstance(log_path, Path):
            new_log_path = log_path
        
        if not new_log_path.exists():
            new_log_path.mkdir(parents=True, exist_ok=True)

        log_file: Path = new_log_path / DEFAULT_FILENAME

        self.stream_handler: StreamHandler = StreamHandler()
        self.file_handler: FileHandler = FileHandler(log_file)
        formatter: Formatter = Formatter(fmt=logfmt, datefmt=datefmt)

        handlers: list[tuple[str, Handler]] = [
            ("file_level", self.file_handler), 
            ("stream_level", self.stream_handler)
        ]
        for level_key, hdlr in handlers:
            hdlr.setFormatter(formatter)
            hdlr.setLevel(levels.get(level_key, DEBUG))

            self.addHandler(hdlr)

        self.setLevel(levels.get("log_level", DEBUG))
    
    def set_logger(self) -> None:
        '''Sets the logger into the class.'''
        setLoggerClass(Log)