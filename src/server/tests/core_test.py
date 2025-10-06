from pathlib import Path
from tests.vars import TestVars as Vars
from system.zipper import Zip, PathArgs
from zipfile import ZipFile
import tests.utils as u

def test_create_zip(tmp_path: Path):
    arm_binary: str = "macdeploy"
    x86_binary: str = "x86_64-macdeploy"
    opt_paths: PathArgs = {
        'arm_binary': arm_binary,
        'x86_binary': x86_binary,
    }

    test_dist_path: Path = tmp_path / "dist"
    test_dist_path.mkdir()

    files: list[str] = [
        "test.pkg", "example.pkg", "item.pkg",
        arm_binary, x86_binary, "folder1/twice.pkg",
        "folder1/folder2/thrice.pkg"
    ]

    for file in files:
        temp_file: Path = test_dist_path / file
        # creating directory if this contains folders.
        has_folders: bool = temp_file.parent.absolute() != test_dist_path.absolute()

        if has_folders:
            temp_file.parent.mkdir(parents=True, exist_ok=True)

        temp_file.touch()

    zip_path: Path = tmp_path / "deploy-test.zip"
    zipper: Zip = Zip(zip_path, u.get_log(str(tmp_path)), path_args=opt_paths)

    status, _ = zipper.start_zip(dist_path=test_dist_path)

    assert status

    zip_obj: ZipFile = ZipFile(zip_path)
    zip_created: bool = zip_path.exists() and len(zip_obj.filelist) == len(files) and status

    assert zip_created