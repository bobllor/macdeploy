from pathlib import Path
from system.zipper import Zip, BinaryArgs
from logger import LogLevelOptions
from zipfile import ZipFile
from typing import Any
from system.vars import Vars
from . import t_utils as ttils
import os
import system.utils as servu

ARM_BINARY: str = ttils.BIN_ARGS["arm"]
X86_BINARY: str = ttils.BIN_ARGS["x86_64"]

def test_create_zip(tmp_path: Path):
    files: list[str] = [
        "test.pkg", "example.pkg", "item.pkg",
        ARM_BINARY, X86_BINARY, "folder1/twice.pkg",
        "folder1/folder2/thrice.pkg"
    ]

    test_dist_path: Path = tmp_path / "dist"
    test_dist_path.mkdir()

    ttils.setup(test_dist_path, files=files, overwrite=True)

    zip_path: Path = tmp_path / Vars.ZIP_FILE_NAME.value

    log_options: LogLevelOptions = ttils.LOG_OPTIONS.copy()
    log_options["log_level"] = 50
    zipper: Zip = Zip(zip_path, ttils.get_log(str(tmp_path), levels=log_options), binary_args=ttils.BIN_ARGS)

    zip_data: dict[str, Any] = zipper.start_zip(dist_path=test_dist_path)
    zip_file_content: list[str] = zip_data["files"]["content"]

    zip_file_content = [file[:-1] if file[-1] == "/" else file for file in zip_file_content]

    zip_created: bool = zip_path.exists()
    dist_files: list[str] = servu.get_dir_list(test_dist_path, include_arg_path=True)

    root_path: str = str(test_dist_path.parent)
    for file in dist_files:
        file = file.replace(root_path + "/", "")

        if file not in zip_file_content:
            assert AssertionError(f"{file} not found, contents: {zip_file_content}")

    assert zip_created

def test_update_zip(tmp_path: Path):
    dist_dir: Path = tmp_path / "dist"
    ttils.setup(dist_dir, files=[ARM_BINARY, X86_BINARY])

    zip_path: Path = tmp_path / Vars.ZIP_FILE_NAME.value
    log_options: LogLevelOptions = ttils.LOG_OPTIONS.copy()
    log_options["log_level"] = 10
    zipper: Zip = Zip(zip_path, ttils.get_log(str(tmp_path), levels=log_options), binary_args=ttils.BIN_ARGS)

    zip_data: dict[str, Any] = zipper.start_zip(dist_path=dist_dir)
    base_length: int = zip_data["files"]["size"]

    zip_obj: ZipFile = ZipFile(zip_path)
    
    if len(zip_obj.filelist) != base_length:
        assert len(zip_obj.filelist) == base_length

    (dist_dir / "update expected.pkg").touch()
    sub_dir: Path = dist_dir / "dir_test1"
    sub_dir.mkdir(parents=True, exist_ok=True)

    zip_data = zipper.start_zip(dist_path=dist_dir)
    new_length: int = zip_data["files"]["size"]

    zip_obj: ZipFile = ZipFile(zip_path)

    assert len(zip_obj.filelist) == new_length + base_length

def test_dir_change(tmp_path: Path):
    dist_dir: Path = tmp_path / "dist"
    ttils.setup(dist_dir, files=[ARM_BINARY, X86_BINARY], overwrite=True)
    
    zip_path: Path = tmp_path / Vars.ZIP_FILE_NAME.value
    log_options: LogLevelOptions = ttils.LOG_OPTIONS.copy()
    log_options["log_level"] = 10
    zipper: Zip = Zip(zip_path, ttils.get_log(str(tmp_path), levels=log_options), binary_args=ttils.BIN_ARGS)
    
    os.chdir("/tmp")
    zipper.start_zip(dist_dir)

    log_file: Path = None
    for path in dist_dir.parent.iterdir():
        if path.suffix == ".log":
            log_file = path
    if log_file == None:
        assert log_file != None
    
    with open(log_file, "r") as file:
        content: list[str] = file.readlines()

    changed_directory: bool = False
    LOG_LINE: str = "updated working directory to"
    for line in content:
        line = line.lower()

        # this will need to be changed if the log message is changed.
        # found in: zipper.py | method Zip._zipper
        if LOG_LINE in line:
            changed_directory = True
            break
    
    assert changed_directory