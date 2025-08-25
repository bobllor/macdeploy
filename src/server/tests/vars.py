from pathlib import Path

# working directory is in server/
class TestVars:
    ROOT_PATH: str = str(Path(__file__).parent)

    _DIST_DIR_NAME: str = "dist-test"
    _PKG_DIR_NAME: str = "pkg-files-test"

    _BINARY_NAME: str = "deploy-test.bin"
    _ZIP_NAME: str = "deploy-test.zip"

    DIST_DIR_PATH: str = f"{ROOT_PATH}/{_DIST_DIR_NAME}"

    PKG_DIR_PATH: str = f"{DIST_DIR_PATH}/{_PKG_DIR_NAME}"
    ZIP_FILE_PATH: str = f"{DIST_DIR_PATH}/{_ZIP_NAME}"

    BINARY_FILE_PATH: str = f"{DIST_DIR_PATH}/{_BINARY_NAME}"