from pathlib import Path
from .vars import Vars
from .system_types import LogInfo, KeyInfo
from . import utils
from logger import Log
from datetime import date
from typing import Any
import re

class Process:
    def __init__(self, *, log: Log):
        '''Contains the functions for the server to process requests.
        
        Parameters
        ----------
            log: Log
                The Log object used for logging.
        '''
        self.log: Log = log

        self._log_info_keys: list[str] = [key for key in LogInfo.__annotations__.keys()]
        self._key_info_keys: list[str] = [key for key in KeyInfo.__annotations__.keys()]

    def add_filevault(self, key_info: KeyInfo, keys_dir: Path | str) -> dict[str, Any]:
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
            self.log.error(f"Missing key, got: {[key for key in key_info]}")
            return validation_res

        keys_path: Path = None
        if isinstance(keys_dir, str):
            keys_path = Path(keys_dir)
        elif isinstance(keys_dir, Path):
            keys_path = keys_dir

        key: str = key_info.get("key", None)
        serial: str = key_info.get("serialTag", None)

        try:
            key_entry: Path = keys_path / serial / key
            self.log.debug("Key path: %s", key_entry)

            key_log = f"No key entries found for device {serial}"

            if not key_entry.parent.exists():
                self._create_entry(key_entry) 
                self.log.info(f"Added {serial} with key {key}")
            else:
                # commented regex out to simplify the process
                #regex_str: str = r"^([A-Za-z0-9]{4}-?)+$"
                prev_key: str = ""

                # getting the previous key for logging purposes
                for child in key_entry.parent.iterdir():
                    # i dont know the regex for this above LMFAO
                    prev_key_name: str = child.name.strip("-")
                    #match_obj: re.Match[str] | None = re.match(regex_str, prev_key_name)

                    #if match_obj != None:
                    prev_key = prev_key_name
                        
                # if key is empty then the there are files inside the serial tag that isn't the key.
                if prev_key != "": 
                    key_log = f"Replaced existing key {prev_key} with {key}"

                    self.log.info(f"{key_log}")
                    utils.unlink_children(path=key_entry.parent)
                    self._create_entry(key_entry)
                elif prev_key == key:
                    key_log = "Key already exists in entry"
                    return utils.generate_response(
                        status="success",
                        content=key_log,
                        statusCode=200
                    )
        except Exception:
            self.log.exception("Failed to write key to server")

            return utils.generate_response(
                status="error",
                content="Unknown error occurred on the server",
                statusCode=500
            )

        return utils.generate_response(
            content=key_log,
            statusCode=200
        )
   
    def add_log(self, log_info: LogInfo, log_dir: Path | str) -> dict[str, Any]:
        '''Adds the log file from the client device to the server.
        It returns a dictionary response indicating its status and message.

        The response contains the status, content, and status code of the method.
        '''
        log_path: Path = None
        if isinstance(log_dir, str):
            log_path = Path(log_dir)
        elif isinstance(log_dir, Path):
            log_path = log_dir

        validation_res: dict[str, Any] = self._validate_info(self._log_info_keys, log_info)
        if validation_res["status"] == "error":
            self.log.error(f"Missing key in response: {[key for key in log_info]}")
            return validation_res

        # used for formatting logs into the correct dates for organization
        date_logs_name: str = date.today().strftime("%Y-%m-%d") + "-logs"

        try:
            log_file_path: Path = log_path / date_logs_name / log_info["logFileName"]

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
