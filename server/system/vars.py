import os

class Vars:
    # server meta
    VERSION_FILE: str = ".server_version"
    YAML_CONFIG: str = "config.yaml"

    # default paths
    HOME: str = os.environ["HOME"]
    MAIN_DIR: str = f"{HOME}/macos-deployment"
    SERVER_DIR: str = "server"

    # zip vars
    ZIP_PATH: str = ""
    ZIP_FILE: str = "pkg.zip"
