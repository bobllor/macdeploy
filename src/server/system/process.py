from pathlib import Path
from .vars import Vars
from .system_types import LogInfo
from .utils import unlink_children
from logger import logger
from datetime import date
import re

class Process:
    '''Contains the functions for the server to process requests.
    
    Parameters
    ----------
        pkg_path: Path
            The Path location to the main packages folder. This is defined inside vars.py as PKG_PATH.
    '''
    def __init__(self):
        # no idea what to add here
        pass

    def add_filevault(self, serial: str, key: str) -> str:
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

        key_log = f"No key entries found for device {serial}"

        if not serial_dir.exists():
            self._create_entry(key_entry) 
            logger.info(f"Added {serial} with key {key}")
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
                    
            # if key is empty then the there are files inside the serial tag that isn't the key.
            if prev_key != "": 
               key_log = f"Found existing key {prev_key}"

            logger.info(f"{key_log}")
            unlink_children(path=serial_dir)
            self._create_entry(key_entry)

        logger.info(f"Added key {key}")

        return key_log
   
    def add_log(self, log_info: LogInfo) -> None:
        '''Adds the log file from the client device to the server.
        
        This is not an actual log generated in Python but rather 
        is the log generated from the client.
        '''
        # used for formatting logs into the correct dates for organization
        date_logs_name: str = date.today().strftime("%Y-%m-%d") + "-logs"

        log_path: Path = Path(Vars.LOGS_PATH.value) / date_logs_name
        log_file_path: Path = log_path / log_info["logFileName"]

        if not log_path.exists():
            log_path.mkdir(parents=True)

        log_file_path.touch()

        logger.info(f"Added log {log_info['logFileName']}")

        with open(log_file_path, "w") as file:
            file.write(log_info["body"])
  
    def _create_entry(self, path: Path) -> None:
        '''Creates the given Path object with its parents.'''
        path.mkdir(parents=True)
        path.touch()
