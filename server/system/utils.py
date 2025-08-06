from pathlib import Path
from .types import FileTree
from typing import Any

def get_dir_dict(path: Path, *, data: dict[str, Any] = None) -> FileTree:
    '''Get the contents of a given Path in the form of a dictionary.'''
    data = {} if not data else data 

    for child in path.iterdir():
        if not child.is_dir():
            data[child.name] = ""
        else:
            temp_dict: FileTree = get_dir_dict(child)
            data[child.name] = temp_dict
    
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

def get_file_content(path: Path) -> str:
    '''Reads a file and sends back its content.'''
    with open(path, "r") as file:
        content: str = file.read().strip()

    return content