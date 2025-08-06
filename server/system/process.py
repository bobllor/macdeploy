from pathlib import Path
from .vars import Vars
from .types import FileTree
from .utils import get_dir_dict
import hashlib, zipfile, subprocess, threading

class Process:
    '''Contains the functions for the server to process requests.'''
    def __init__(self, pkg_path: Path):
        self.server_files: FileTree = get_dir_dict(pkg_path)

        self._zipping_lock: threading.Lock = threading.Lock()

        #print(self.server_files)

    def add_filevault(self, serial: str, key: str):
        '''Adds the laptop device and key to the server.

        If there is an existing entry then the contents of the entry
        will be removed and replaced with the new key.
        
        Parameters:
        -----------
            key: str
                The FileVault key generated from the device.
        '''
        serial_dir: Path = Path(f"{Vars.FILEVAULT_PATH.value}/{serial}")
        key_entry: Path = serial_dir / key

        if not serial_dir.exists():
            self._create_entry(key_entry) 
        else:
            # this does not remove the parent directory but its contents.
            print(f"found existing key {serial_dir[0]}")
            self._unlink_children(path=serial_dir)
            self._create_entry(key_entry)
    
    def generate_hash(self, path: Path, *, data: list[str] = None) -> str:
        '''Generates the hash by using the file name given from a Path.
        
        If the Path is a directory, then it will recursively go through its contents
        and generate the hash once finished.

        The file names are **case sensitive**.
        '''
        data = [] if not data else data

        if path.is_dir():
            self._hash_recursion(path, data)
            data.sort()
        else:
            data.append(path.name)
        
        #print(data)

        return hashlib.md5("".join(data).encode()).hexdigest()
    
    def update_zip(self, zip_path: Path) -> None:
        '''Update a given Path to a ZIP file by checking the current packages with
        the packages available in the ZIP file.

        If the ZIP file does not exist, then a new one will be created.
        .'''
        if not zip_path.exists():
            # files to zip: packages folder, go binary
            files_to_zip: str = f"{Vars.PKG_PATH.value}"
            zip_cmd: list[str] = f"zip -f {Vars.ZIP_FILE_NAME.value} {files_to_zip}".split()
            zip_cmd: list[str] = ["powershell", "-c", "sleep 5"] # temp
            threading.Thread(target=self._execute, args=(zip_cmd,)).start()

            return
    
    def _execute(self, cmd: list[str]) -> None:
        '''Runs a subprocess command for execution.
        
        This is blocking by default, run with threading if non-blocking is required.
        '''
        locked_thread: bool = self._zipping_lock.acquire(blocking=False)

        print(locked_thread)

        if locked_thread:
            try:
                print(f"Running command {" ".join(cmd)}")
                subprocess.run(cmd)
                print("Command finished")
            finally:    
                self._zipping_lock.release()
        else:
            print("New execution attempt initiated, blocking")
            return


    def _hash_recursion(self, path: Path, data: list[str]) -> None:
        '''Helper function for `self.generate_hash`, recursively iterates through a
        directory and mutates the list in-place.
        '''
        data.append(path.name)

        for child in path.iterdir():
            if child.is_dir():
                self._hash_recursion(child, data=data)
            else:
                data.append(child.name)

    def _create_entry(self, path: Path) -> None:
        '''Creates the given Path object with its parents.'''
        path.mkdir(parents=True)
        path.touch()

    def _unlink_children(self, path: Path) -> None:
        '''Removes all children from a given Path'''
        if not path.is_dir():
            path.unlink()
            return

        for file in path.iterdir():
            if file.is_dir():
                self._unlink_children(file)
                file.rmdir()