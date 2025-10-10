from logger import Log, LogLevelOptions
from logging import DEBUG, CRITICAL
from system.zipper import BinaryArgs
from system.vars import Vars
from pathlib import Path
from zipfile import ZipFile

LOG_OPTIONS: LogLevelOptions = {
    "log_level": CRITICAL
}
BIN_ARGS: BinaryArgs = {
    "arm": Vars.ARM_BINARY_NAME.value,
    "x86_64": Vars.X86_BINARY_NAME.value,
}

def get_log(path: str, *, levels: LogLevelOptions = None) -> Log:
    if not levels:
        levels = {"log_level": DEBUG}

    return Log(log_path=path, levels=levels)

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