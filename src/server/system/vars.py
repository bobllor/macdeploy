from pathlib import Path
from enum import Enum

class Vars(Enum):
    ROOT_PATH = str(Path(__file__).parents[3]) # ensures we are working in the root directory
    
    # file names
    ZIP_FILE_NAME = "deploy.zip"
    ARM_BINARY_NAME = "macdeploy"
    X86_BINARY_NAME = "x86_64-macdeploy"

    # directory names
    _KEYS_DIR_NAME = "keys"
    _SERVER_DIR_NAME = "server"
    _LOGS_DIR_NAME = "logs"
    _SERVER_LOG_NAME = "server-logs"

    # default directory paths
    KEYS_PATH = f"{ROOT_PATH}/{_KEYS_DIR_NAME}"

    SERVER_PATH = f"{ROOT_PATH}/src/{_SERVER_DIR_NAME}"
    DIST_PATH = f"{ROOT_PATH}/dist" # binary and pkg-files are located in here
    ZIP_PATH = f"{ROOT_PATH}/build" # zip is located in build

    LOGS_PATH = f"{ROOT_PATH}/{_LOGS_DIR_NAME}"
    SERVER_LOG_PATH = f"{LOGS_PATH}/{_SERVER_LOG_NAME}"