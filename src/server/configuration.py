from pathlib import Path

ROOT_PATH: Path = Path(__file__).parents[2] # ensures we are working in the root directory

# file names
ZIP_NAME: str = "deploy.zip"
ARM_BINARY_NAME: str = "macdeploy"
X86_BINARY_NAME: str = "x86_64-macdeploy"

# directories
KEYS_NAME: str= "keys"
SERVER_NAME: str = "server"
LOGS_NAME: str = "logs"
SERVER_LOGS_NAME: str = "server-logs"
ZIP_DIR_NAME: str = "zip-build"
DIST_DIR_NAME: str = "dist"

# directory paths
KEYS_PATH: Path = ROOT_PATH / KEYS_NAME

SERVER_PATH: Path = ROOT_PATH / "src" / SERVER_NAME
# client files, binaries, are stored in this location
DIST_PATH: Path = ROOT_PATH / DIST_DIR_NAME
# zip file is stored in this location
ZIP_PATH = ROOT_PATH / ZIP_DIR_NAME

LOGS_PATH: Path = ROOT_PATH / LOGS_NAME
SERVER_LOGS_PATH: Path = LOGS_PATH / SERVER_LOGS_NAME