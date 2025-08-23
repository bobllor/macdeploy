from pathlib import Path
from system.vars import Vars

filevault_dir: str = Vars.FILEVAULT_PATH.value
logs_dir: str = Vars.LOGS_PATH.value
pkg_dir: str = Vars.PKG_PATH.value
server_log_dir: str = Vars.SERVER_LOG_PATH.value

dir_list: list[str] = [filevault_dir, logs_dir, pkg_dir, server_log_dir]

# these directories are checked during runtime as well.
for ele in dir_list:
    path: Path = Path(ele)

    if not path.exists():
        path.mkdir()