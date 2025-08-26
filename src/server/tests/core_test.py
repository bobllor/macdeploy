from pathlib import Path
from tests.vars import TestVars as Vars
from system.zipper import Zip, PathArgs
from zipfile import ZipFile

def test_zip_update():
    zip_path: Path = Path(Vars.ZIP_FILE_PATH)
    opt_paths: PathArgs = {
        'arm_binary': Vars.BINARY_FILE_PATH,
        'x86_binary': "deploy-x86-test.bin",
    }

    zipper: Zip = Zip(zip_path, path_args=opt_paths)

    status, _ = zipper.start_zip(dist_path=Path(Vars.DIST_DIR_PATH))
    zip_obj: ZipFile = ZipFile(zip_path)

    zip_created: bool = zip_path.exists() and len(zip_obj.filelist) > 0
    #zip_path.unlink()

    assert zip_created