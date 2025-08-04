import os
from pathlib import Path
from enum import Enum

class Vars(Enum):
    HOME = os.environ["HOME"]
    MAIN_PATH = Path(__file__).parents[2] # ensures we are working in the main directory

    # server meta
    PKG_HASH = ".pkg_metadata"
    YAML_CONFIG = "config.yaml"
    
    # file names
    ZIP_FILE_NAME = "pkgs.zip"

    # directory names
    FILEVAULT_DIR_NAME = "filevault-keys"
    PKG_DIR_NAME = "pkg-files"
    SERVER_DIR_NAME = "server"
    LOG_DIR_NAME = "logs"
    ZIP_DIR_NAME = ""

    # default directory paths
    FILEVAULT_PATH = f"{MAIN_PATH}/{FILEVAULT_DIR_NAME}"
    PKG_PATH = f"{MAIN_PATH}/{PKG_DIR_NAME}"
    SERVER_PATH = "server"
    ZIP_PATH = ""