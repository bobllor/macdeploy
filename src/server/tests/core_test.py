from pathlib import Path
from tests.vars import TestVars as Vars
from system.zipper import Zip, PathArgs
from zipfile import ZipFile

def test_zip_update():
    zip_path: Path = Path(Vars.ZIP_FILE_PATH)
    opt_paths: PathArgs = {
        'binary_file': Vars.BINARY_FILE_PATH,
        'dist_dir': Vars.DIST_DIR_PATH,
    }

    zipper: Zip = Zip(zip_path, path_args=opt_paths)

    pkg_path: Path = Path(Vars.PKG_DIR_PATH)

    zipper.start_zip(pkg_path, dist_dir=opt_paths['dist_dir'])
    zip_obj: ZipFile = ZipFile(zip_path)

    zip_created: bool = zip_path.exists() and len(zip_obj.filelist) > 0
    #zip_path.unlink()

    assert zip_created