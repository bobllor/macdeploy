from pathlib import Path
from .vars import Vars
from .types import Flags
from .utils import get_dir_list, unlink_children
import hashlib, zipfile, subprocess, threading, re

class Process:
    '''Contains the functions for the server to process requests.
    
    Parameters
    ----------
        pkg_path: Path
            The Path location to the main packages folder. This is defined inside vars.py as PKG_PATH.
    '''
    def __init__(self, pkg_path: Path):
        self.server_files: list[str] = get_dir_list(pkg_path, replace_home=True)
        self._zipping_lock: threading.Lock = threading.Lock()

        self.flags: Flags = {
            "zip_status": True
        }

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
            regex_str: str = r"^([A-Za-z0-9]{4}-?)+$"
            prev_key: str = ""

            # getting the previous key for logging purposes
            for child in serial_dir.iterdir():
                # i dont know the regex for this above LMFAO
                prev_key_name: str = child.name.strip("-")
                match_obj: re.Match[str] | None = re.match(regex_str, prev_key_name)

                if match_obj != None:
                    prev_key = prev_key_name
                    break
            
            key_log: str = f"Found existing key {prev_key}"
            if prev_key == "": 
                key_log = f"Error finding key in {serial} directory"

            print(key_log)
            unlink_children(path=serial_dir)
            self._create_entry(key_entry)

        print(f"Added key {key}")
    
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
        '''
        if not zip_path.exists():
            # files to zip: packages folder, go binary
            files_to_zip: str = f"{Vars.PKG_PATH.value}"
            zip_cmd: list[str] = f"zip -f {Vars.ZIP_FILE_NAME.value} {files_to_zip}".split()

            threading.Thread(target=self._execute, args=(zip_cmd,)).start()

            return
    
    def _execute(self, cmd: list[str]) -> None:
        '''Runs a subprocess command for execution.
        
        This is blocking by default, run with threading if non-blocking is required.
        '''
        self.flags["zip_status"] = False
        locked_thread: bool = self._zipping_lock.acquire(blocking=False)

        print(locked_thread)

        if locked_thread:
            try:
                print(f"Running command {" ".join(cmd)}")
                subprocess.run(cmd)
                print("Command finished")
            finally:    
                self._zipping_lock.release()
                self.flags["zip_status"] = True
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
    
    @property
    def flags(self) -> Flags:
        '''The status flags on operations.
        By default all flags are True, and under certain operations will they be False.

        This is used for conditionals for the endpoints.
        '''
        return self.flags