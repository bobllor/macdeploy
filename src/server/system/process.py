from pathlib import Path
from .vars import Vars
from .system_types import LogInfo, KeyInfo
from . import utils
from logger import Log
from datetime import date
from typing import Any
import re

class Process:
    def __init__(self, *, log_dir: Path, log: Log):
        '''Contains the functions for the server to process requests.
        
        Parameters
        ----------
            log_dir: Path
                The directory where logging is stored.

            log: Log
                The Log object used for logging.
        '''
        self.log_dir: Path = log_dir
        self.log: Log = log

        self._log_info_keys: list[str] = [key for key in LogInfo.__annotations__.keys()]
        self._key_info_keys: list[str] = [key for key in KeyInfo.__annotations__.keys()]

    def add_filevault(self, key_info: KeyInfo, keys_dir: Path = Path(Vars.FILEVAULT_PATH.value)) -> dict[str, Any]:
        '''Adds the laptop device and key to the server.

        If there is an existing entry then the contents of the entry
        will be removed and replaced with the new key.
        
        Parameters:
        -----------
            key: str
                The FileVault key generated from the device.
        '''
        validation_res: dict[str, Any] = self._validate_info(self._key_info_keys, key_info)
        if validation_res["status"] == "error":
            return validation_res

        key: str = key_info.get("key", None)
        serial: str = key_info.get("serialTag", None)

        try:
            serial_dir: Path = keys_dir / serial
            key_entry: Path = serial_dir / key

            key_log = f"No key entries found for device {serial}"

            if not serial_dir.exists():
                self._create_entry(key_entry) 
                self.log.info(f"Added {serial} with key {key}")
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

                    self.log.info(f"{key_log}")
                    utils.unlink_children(path=serial_dir)
                    self._create_entry(key_entry)
        except Exception:
            self.log.info("Failed to write key to server")

            return utils.generate_response(
                status="error",
                content="Unknown error occurred on the server",
                statusCode="500"
            )

        self.log.info(f"Added key {key}")

        return utils.generate_response(
            content=key_log,
            statusCode=200
        )
   
    def add_log(self, log_info: LogInfo) -> dict[str, Any]:
        '''Adds the log file from the client device to the server.
        It returns a dictionary response indicating its status and message.

        The response contains the status, content, and status code of the method.
        '''
        validation_res: dict[str, Any] = self._validate_info(self._log_info_keys, log_info)
        if validation_res["status"] == "error":
            return validation_res

        # used for formatting logs into the correct dates for organization
        date_logs_name: str = date.today().strftime("%Y-%m-%d") + "-logs"

        try:
            log_file_path: Path = self.log_dir / date_logs_name / log_info["logFileName"]

            if not log_file_path.parent.exists():
                log_file_path.parent.mkdir(parents=True, exist_ok=True)

            with open(log_file_path, "w") as file:
                file.write(log_info["body"])
        except Exception:
            self.log.exception("Failed to write log to the server")

            return utils.generate_response(
                status="error",
                content="An unknown error occurred on the server",
                statusCode=500
            )
        
        self.log.info(f"Added log {log_info['logFileName']}")
        self.log.debug(f"Log location: {log_file_path}")

        return utils.generate_response(
            content="Successfully added logs to server",
            statusCode=200
        )
    
    def _validate_info(self, keys_to_check: list[str], info: dict[str, Any]) -> dict[str, Any]:
        '''Checks the info dictionary for validating the responses.'''
        for key in keys_to_check:
            if key not in info:
                return utils.generate_response(
                    status="error", 
                    content=f"Missing key in response",
                    statusCode=400
                )
            
            if info[key] == "":
                return utils.generate_response(
                    status="error", 
                    content=f"No content given for key",
                    statusCode=400
                )

        return utils.generate_response(statusCode=200, content="Successful validation")
  
    def _create_entry(self, path: Path) -> None:
        '''Creates the given Path object with its parents.'''
        path.mkdir(parents=True)
        path.touch()
