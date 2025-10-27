from pathlib import Path
from configuration import KEYS_PATH, LOGS_PATH

dir_list: list[Path] = [KEYS_PATH, LOGS_PATH]

# these directories are checked during runtime as well.
for path in dir_list:
    if not path.exists():
        path.mkdir(parents=True)