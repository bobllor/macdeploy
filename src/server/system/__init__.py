from pathlib import Path
from system.vars import Vars

filevault_dir: str = Vars.KEYS_PATH.value
logs_dir: str = Vars.LOGS_PATH.value

dir_list: list[str] = [filevault_dir, logs_dir]

# these directories are checked during runtime as well.
for ele in dir_list:
    path: Path = Path(ele)

    if not path.exists():
        path.mkdir(parents=True)