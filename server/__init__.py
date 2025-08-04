from pathlib import Path
from system.vars import Vars

filevault_dir: str = f"{Vars.FILEVAULT_PATH.value}/{Vars.FILEVAULT_DIR_NAME.value}"
logs_dir: str = f"{Vars.MAIN_PATH}/{Vars.LOG_DIR_NAME}"

dir_list: list[str] = [filevault_dir, logs_dir]

# these directories are checked during runtime as well.
for ele in dir_list:
    path: Path = Path(dir)

    if not path.exists():
        path.mkdir()