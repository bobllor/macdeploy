from typing import TypedDict
from logger import LogLevelOptions, Log
from pathlib import Path

class Config(TypedDict):
    zip_path: Path | str
    log_path: Path | str
    log_server_path: Path | str
    keys_path: Path | str
    token_path: Path | str
    dist_path: Path | str
    testing: bool
    log_levels: LogLevelOptions
    token_bits: int