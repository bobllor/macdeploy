from system.process import Process
from pathlib import Path
from logger import Log, LogLevelOptions
from system.system_types import LogInfo, KeyInfo
from . import t_utils as ttils
from datetime import datetime
from typing import Any
import system.utils as utils

def test_add_log(tmp_path: Path):
    log: Log = ttils.get_log(tmp_path)
    process: Process = Process(log=log)

    body: str = "Log content\nLine one\nLine two\nLine three"
    log_file: str = datetime.now().strftime("%Y-%m-%dT%H-%M-%S.SERIAL.log")

    log_content: LogInfo = {
        "body": body,
        "logFileName": log_file
    } 

    log_res: dict[str, Any] = process.add_log(log_content, tmp_path)
    if log_res["status"] == "error":
        assert log_res["status"] != "error"

    files: list[str] = utils.get_dir_list(tmp_path)
    log_path: Path = Path("")

    for file in files:
        path: Path = Path(file)

        if path.suffix == ".log" and "server" not in path.name:
            log_path = path
            break
    
    with open(log_path, "r") as file:
        content: str = file.read().strip()

    assert content == body

def test_multiple_keys_log(tmp_path: Path):
    log: Log = ttils.get_log(tmp_path)
    process: Process = Process(log=log)

    body: str = "Log content\nLine one\nLine two\nLine three"
    log_file: str = datetime.now().strftime("%Y-%m-%dT%H-%M-%S.SERIAL2.log")

    log_content: LogInfo = {
        "body": body,
        "logFileName": log_file,
        "extraKey1": "test",
        "extraKey2": "test again",
    } 

    log_res: dict[str, Any] = process.add_log(log_content, tmp_path)
    if log_res["status"] == "error":
        assert log_res["status"] != "error"

    files: list[str] = utils.get_dir_list(tmp_path)
    log_path: Path = Path("")

    for file in files:
        path: Path = Path(file)

        if path.suffix == ".log" and "server" not in path.name:
            log_path = path
            break
    
    with open(log_path, "r") as file:
        content: str = file.read().strip()

    assert content == body

def test_fail_empty_string(tmp_path: Path):
    log: Log = ttils.get_log(tmp_path)
    process: Process = Process(log=log)

    log_file: str = datetime.now().strftime("%Y-%m-%dT%H-%M-%S.SERIAL.log")

    log_content: LogInfo = {
        "body": "",
        "logFileName": log_file
    } 

    log_res: dict[str, Any] = process.add_log(log_content, tmp_path)

    assert log_res["status"] == "error"

def test_fail_wrong_key(tmp_path: Path):
    log: Log = ttils.get_log(tmp_path)
    process: Process = Process(log=log)

    body: str = "Log content\nLine one\nLine two\nLine three"
    log_file: str = datetime.now().strftime("%Y-%m-%dT%H-%M-%S.SERIAL.log")

    log_content: LogInfo = {
        "body": body,
        "fakeFileBTW": log_file
    } 

    log_res: dict[str, Any] = process.add_log(log_content, tmp_path)

    assert log_res["status"] == "error"