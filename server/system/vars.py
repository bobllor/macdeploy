import os
from pathlib import Path
from enum import Enum

class Vars(Enum):
    HOME = os.environ["HOME"]
    _MAIN_PATH = Path(__file__).parents[2] # ensures we are working in the main directory
    
    # file names
    ZIP_FILE_NAME = "deploy.zip"
    BINARY_NAME = "deploy"
    YAML_CONFIG = "config.yaml"

    # directory names
    _FILEVAULT_DIR_NAME = "filevault-keys"
    _PKG_DIR_NAME = "pkg-files"
    _SERVER_DIR_NAME = "server"
    _LOGS_DIR_NAME = "logs"
    _ZIP_DIR_NAME = "deploy-zip"
    _SERVER_LOG_NAME = "server-logs"

    # default directory paths
    FILEVAULT_PATH = f"{_MAIN_PATH}/{_FILEVAULT_DIR_NAME}"
    PKG_PATH = f"{_MAIN_PATH}/{_PKG_DIR_NAME}"
    SERVER_PATH = f"{_MAIN_PATH}/{_SERVER_DIR_NAME}"
    ZIP_PATH = f"{_MAIN_PATH}/{_ZIP_DIR_NAME}"
    LOGS_PATH = f"{_MAIN_PATH}/{_LOGS_DIR_NAME}"
    SERVER_LOG_PATH = f"{LOGS_PATH}/{_SERVER_LOG_NAME}"