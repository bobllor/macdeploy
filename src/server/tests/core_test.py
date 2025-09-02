from pathlib import Path
from tests.vars import TestVars as Vars
from system.zipper import Zip, PathArgs
from zipfile import ZipFile

def test_zip_update(tmp_path):
    server_root: Path = Path(Vars.ROOT_PATH).parents[2]
    zip_path: Path = server_root / "deploy-test.zip"

    opt_paths: PathArgs = {
        'arm_binary': Vars.BINARY_FILE_PATH,
        'x86_binary': "deploy-x86-test.bin",
    }

    test_dist_path: Path = tmp_path / "dist-test"
    test_dist_path.mkdir()

    for path, _, files in Path(Vars.DIST_DIR_PATH).walk():
        new_test_path: Path = test_dist_path / path.name if path.name != "dist-test" else test_dist_path

        for file in files:
            if not new_test_path.exists():
                new_test_path.mkdir()
            
            final_path: Path = new_test_path / file
            final_path.touch()

    zipper: Zip = Zip(zip_path, path_args=opt_paths)

    status, _ = zipper.start_zip(dist_path=test_dist_path)
    zip_obj: ZipFile = ZipFile(zip_path)

    zip_created: bool = zip_path.exists() and len(zip_obj.filelist) > 0 and status
    zip_path.unlink()

    assert zip_created