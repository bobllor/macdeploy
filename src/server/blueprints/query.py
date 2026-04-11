from flask import Blueprint, jsonify
from logger import Log
from logger import Log
from app_types import Config
from pathlib import Path
from typing import Any, TypedDict, Literal
from datetime import datetime, timezone
import system.utils as utils
import os


class FileData(TypedDict):
    name: str
    modified: datetime
    size: int

class Query():
    def __init__(self, *, config: Config, logger: Log):
        '''Processing class blueprint.
        
        Parameters
        ----------
            config: Config
                A dictionary of configuration settings.
        '''
        self.logger: Log = logger
        self.config: Config = config

    def get_blueprint(self) -> Blueprint:
        bp: Blueprint = Blueprint("query", __name__)

        @bp.get("/api/devices/<device>")
        def get_device_info(device: str):
            '''Get the metadata of the device and its filevault key, if either exists.
            The file information contains its metadata information.
            Both files will be appended to the `content` array of the response.
            '''
            keys_path: Path = self.config["keys_path"]
            if not keys_path.exists():
                self.logger.warning(f"Created missing 'keys' folder: {keys_path}")
                keys_path.mkdir(parents=True, exist_ok=True)

            # casing does not matter with Path 
            device = device.strip()
            # could maybe rewrite this into SQL now that i know it... nah maybe not - 4/8/2026
            device_path: Path = keys_path / device

            content: list[FileData] = []
            res: dict[str, Any] = utils.generate_response("success", status_code=200, content=content, message="Device found")

            if device_path.exists():
                stat: os.stat_result = device_path.stat()

                device_file_data: FileData = {
                    "name": device,
                    "modified": datetime.fromtimestamp(stat.st_mtime, timezone.utc),
                    "size": stat.st_size,
                }

                content.append(device_file_data)

                key_path: Path | None = None
                for child in device_path.iterdir():
                    file_name: str = child.name

                    # its expected the file has no suffix, this gets
                    # any cases where it does have a suffix.
                    # do note that this only cleans the ending section, the
                    # actual pattern must still match.
                    if child.suffix != "":
                        file_name = file_name.split(".")[0]

                    if utils.is_filevault_key(file_name):
                        key_path = child
                        break
                    
                if key_path is not None:
                    key_stat: os.stat_result = key_path.stat()

                    key_data: FileData = {
                        "name": key_path.name,
                        "modified": datetime.fromtimestamp(key_stat.st_mtime, timezone.utc),
                        "size": stat.st_size,
                    } 

                    content.append(key_data)
            else:
                self.logger.warning(f"Device '{device}' does not exist in entries")
                res["status"] = "error"
                res["message"] = f"Device {device} is not found"
        
            return res, res["status_code"]

        return bp