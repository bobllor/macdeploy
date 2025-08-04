from pathlib import Path
from vars import Vars

class Process:
    def __init__(self, serial: str = "UNKNOWN"):
        self.serial: str = serial

    def add_filevault(self, key: str):
        '''Adds the laptop device and key to the server.
        
        If there is an existing entry then the contents of the entry
        will be removed and replaced with the new key.
        
        Parameters:
        -----------
            key: str
                The FileVault key generated from the device.
        '''
        serial_dir: Path = Path(f"{Vars.FILEVAULT_PATH.value}/{self.serial}")
        key_entry: Path = serial_dir / key

        if not serial_dir.exists():
            self._create_entry(key_entry) 
        else:
            self._unlink_children(path=serial_dir)
            self._create_entry(key_entry)
        
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