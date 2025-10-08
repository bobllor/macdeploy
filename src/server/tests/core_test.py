from pathlib import Path
from system.zipper import Zip, PathArgs
from logger import LogLevelOptions
from logging import CRITICAL
from zipfile import ZipFile
from typing import Any
from . import t_utils as ttils
import os
import system.utils as servu

ARM_BINARY: str = "macdeploy"
X86_BINARY: str = "x86_64-macdeploy"
ZIP_FILE: str = "deploy.zip"
OPT_PATHS: PathArgs = {
    'arm_binary': ARM_BINARY,
    'x86_binary': X86_BINARY,
}

def test_create_zip(tmp_path: Path):
    files: list[str] = [
        "test.pkg", "example.pkg", "item.pkg",
        ARM_BINARY, X86_BINARY, "folder1/twice.pkg",
        "folder1/folder2/thrice.pkg"
    ]

    test_dist_path: Path = tmp_path / "dist"
    test_dist_path.mkdir()

    setup(test_dist_path, files=files, overwrite=True)

    zip_path: Path = tmp_path / ZIP_FILE

    log_options: LogLevelOptions = ttils.LOG_OPTIONS.copy()
    log_options["log_level"] = 50
    zipper: Zip = Zip(zip_path, ttils.get_log(str(tmp_path), levels=log_options), path_args=OPT_PATHS)

    zip_data: dict[str, Any] = zipper.start_zip(dist_path=test_dist_path)
    zip_file_content: list[str] = zip_data["files"]["content"]

    zip_file_content = [file[:-1] if file[-1] == "/" else file for file in zip_file_content]

    zip_created: bool = zip_path.exists()
    dist_files: list[str] = servu.get_dir_list(test_dist_path, include_arg_path=True)

    root_path: str = str(test_dist_path.parent)
    for file in dist_files:
        file = file.replace(root_path + "/", "")

        if file not in zip_file_content:
            print(f"{file} not found, contents: {zip_file_content}")
            assert False

    assert zip_created

def test_update_zip(tmp_path: Path):
    dist_dir: Path = tmp_path / "dist"
    setup(dist_dir, files=[ARM_BINARY, X86_BINARY])

    zip_path: Path = tmp_path / ZIP_FILE
    log_options: LogLevelOptions = ttils.LOG_OPTIONS.copy()
    log_options["log_level"] = 10
    zipper: Zip = Zip(zip_path, ttils.get_log(str(tmp_path), levels=log_options), path_args=OPT_PATHS)

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
    setup(dist_dir, files=[ARM_BINARY, X86_BINARY], overwrite=True)
    
    zip_path: Path = tmp_path / ZIP_FILE
    log_options: LogLevelOptions = ttils.LOG_OPTIONS.copy()
    log_options["log_level"] = 10
    zipper: Zip = Zip(zip_path, ttils.get_log(str(tmp_path), levels=log_options), path_args=OPT_PATHS)
    
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
        print(line)
        if LOG_LINE in line:
            changed_directory = True
            break
    
    assert changed_directory

def setup(path: Path, *, files: list[str] = None, overwrite: bool = False):
    '''Setups the directories with the files.
    
    Parameters
    ----------
        path: Path
            The default directory that the files will be made in.
        
        files: list[str], default None
            A list of **relative** file paths. These files are appended to the `path` argument.
            Default files exist, if files are given then it will append to the default files.
            To utilize a custom file, include `overwrite = True`. 
        
        overwrite: bool, default False
            Boolean status indicating to overwrite the files or not. By default it is false,
            indicating that any `files` arguments will append to the default files defined in
            the function. Otherwise, overwrite it.
    '''
    default_files: list[str] = [
        "folder1/subfolder1/absolute.pkg", "bullet.pkg",
        "animal.pkg", "folder2/subfolder1/timezone.app",
        "folder2/subfolder2/manatee.txt", "apricot.txt",
        "folder2/treasure"
    ]

    if overwrite:
        default_files = files
    else:
        default_files.extend(files)

    for file in default_files:
        temp_file: Path = path / file
        # creating directory if this contains folders.
        has_folders: bool = temp_file.parent.absolute() != path.absolute()

        if has_folders or temp_file.suffix == "":
            temp_file.parent.mkdir(parents=True, exist_ok=True)

        temp_file.touch()