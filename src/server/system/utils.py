from pathlib import Path
from flask import make_response, Response
from configuration import ROOT_PATH

def get_dir_list(path: Path | str, data: list[str] = None, 
    *, replace_root: bool = False, include_arg_path: bool = False) -> list[str]:
    '''Get the contents of a given directory Path in the form of a list of strings.
    The list contains all files and directories, but it *does not include the argument path directory*.
    
    Parameters
    ----------
        path: Path | str
            The directory being searched in, it can be a Path or a path string.
        
        data: list[str], default None
            List of paths, if a list is given then that list will get extended with the files.
            Otherwise, it is default None and returns a new list.
        
        replace_root: bool, default False
            Replaces the project root path from the absolute path of the paths.
        
        include_arg_path: bool, default False
            Includes the given path argument with the list of paths on the return value.
    '''
    data = [] if not data else data 
    if include_arg_path:
        file: str = str(path)
        if replace_root: file = file.replace(str(ROOT_PATH) + "/", "")
        data.append(file)

    if isinstance(path, str):
        path = Path(path)

    for child in path.iterdir():
        file: str = str(child)
        if not child.is_dir():
            if replace_root:
                # removing the home path and the trailing slash
                file = file.replace(str(ROOT_PATH) + "/", "")

            data.append(file)
        else:
            if replace_root:
                # removing the home path and the trailing slash
                file = file.replace(str(ROOT_PATH) + "/", "")

            data.append(file)
            temp_list: list[str] = get_dir_list(child, replace_root=replace_root)
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

def generate_response(status: str = "success", **kwargs) -> dict[str, str]:
    '''Generates a dictionary for a response.
    
    Parameters
    ----------
        status: str, default *"success"*
            The status of the response. By default it is "success".
    '''
    response: dict[str, str] = {
        "status": status
    }

    for key, value in kwargs.items():
        response[key] = value

    return response

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

def write_to_file(file_path: Path | str, content: str) -> None:
    '''Writes to a file from a given path.'''
    with open(file_path, "w") as file:
        file.write(content)
    
def read_from(file_path: Path | str) -> str:
    with open(file_path, "r") as file:
        content: str = file.read()

    return content