from pathlib import Path
from .types import FileTree
from .vars import Vars
from typing import Any
from zipfile import ZipFile, ZipInfo
import io

def get_dir_list(path: Path, data: list[str] = None, *, replace_home: bool = False) -> list[str]:
    '''Get the contents of a given directory Path in the form of a list of strings.'''
    data = [] if not data else data 

    for child in path.iterdir():
        if not child.is_dir():
            file: str = str(child)

            if replace_home:
               file = file.replace(str(Vars._MAIN_PATH.value) + "/", "")

            data.append(file)
        else:
            temp_list: list[str] = get_dir_list(child, replace_home=replace_home)
            data.extend(temp_list)
    
    return data

def mk_paths(paths: list[Path], *, mk_dir: bool = True):
    '''Makes a file or directory if it does not exist from a given list of Paths.
    
    If the file does not end in an extension, then it will be assumed it is a directory.
    If `mk_dir` is False, then all Paths will be made as a file instead. 
    '''
    for path in paths:
        if not path.exists():
            ext: str = path.suffix
            if not mk_dir or ext != '':
                path.touch()
            else:
                path.mkdir(parents=True)

def unlink_children(path: Path) -> None:
    '''Removes all children from a given Path.
    
    This does not remove the given Path.
    '''
    if not path.is_dir():
        path.unlink()
        return

    for file in path.iterdir():
        if file.is_dir():
            unlink_children(file)
            file.rmdir()
        else:
            file.unlink()