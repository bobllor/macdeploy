from flask.testing import FlaskClient
from werkzeug.test import TestResponse
from configuration import ZIP_NAME
from pathlib import Path
from system.system_types import LogInfo, KeyInfo
from system.zipper import Zip
from zipfile import ZipFile
from typing import Any
from logger import Log
from . import t_utils as ttils
from threading import Thread
import system.utils as utils
import json

def test_add_key(tmp_path: Path, client: FlaskClient):
    serial: str = "SERIAL1234"
    key: str = "1235-6789-1024-ABC0"

    key_info: KeyInfo = {
        "key": key,
        "serialTag": serial,
    } 

    response: TestResponse = client.post("/api/fv", json=key_info)
    if response.status_code != 200:
        raise AssertionError(f"Got {response.status_code}: {response.data.decode()}")

    files: list[str] = utils.get_dir_list(tmp_path)
    for file in files:
        path: Path = Path(file)

        if path.name == key:
            assert True

def test_add_existing_key(tmp_path: Path, client: FlaskClient):
    serial: str = "SERIAL1234"
    key: str = "1235-6789-1024-ABC0"

    key_info: KeyInfo = {
        "key": key,
        "serialTag": serial,
    } 

    response: TestResponse = client.post("/api/fv", json=key_info)
    if response.status_code != 200:
        raise AssertionError(f"Got {response.status_code}: {response.data.decode()}")

    new_key: str = "A-REPLACED-KEY-HERE"
    key_info["key"] = new_key

    response: TestResponse = client.post("/api/fv", json=key_info)
    if response.status_code != 200:
        raise AssertionError(f"Got {response.status_code}: {response.data.decode()}")

    files: list[str] = utils.get_dir_list(tmp_path)

    found_path: Path = Path("nonexistentpath")
    found_key: bool = False
    for file in files:
        file_path: Path = Path(file)

        if file_path.name == new_key:
            found_key = True
            found_path = file_path
            break
    
    # if the loop assertion fails above.
    assert found_key and found_path.exists()

def test_add_key_fail(client: FlaskClient):
    key: str = "1235-6789-1024-ABC0"
    key_info: KeyInfo = {
        "key": key,
    } 

    response: TestResponse = client.post("/api/fv", json=key_info)
    if response.status_code == 200:
        raise AssertionError(f"Got {response.status_code}: {response.data.decode()}")

    content: dict[str, Any] = json.loads(response.data)    

    res_msg: str = content["content"]
    status: str = content["status"]
    expected_msg: str = "missing key in response"

    assert status == "error" and expected_msg in res_msg.lower()

def test_add_log(tmp_path: Path, client: FlaskClient):
    api: str = "/api/log"

    body: str = "line1\nline2\nline3\n\nline4\n"
    file_name: str = "SOMETHING.SERIAL2.log"
    log_info: LogInfo = {
        "body": body,
        "logFileName": file_name
    }

    res: TestResponse = client.post(api, json=log_info)

    files: list[str] = utils.get_dir_list(tmp_path / "logs")

    log_file: Path = None
    for file in files:
        if file_name in file:
            log_file = Path(file)
            break
    
    if log_file is None:
        raise AssertionError(f"Could not find generated log file at {log_file}")
    
    log_content: str = ""
    with open(log_file, "r") as file:
        log_content = file.read()

    assert res.status_code == 200 and log_content == body

def test_add_log_fail(tmp_path: Path, client: FlaskClient):
    api: str = "/api/log"

    file_name: str = "SOMETHING.SERIAL2.log"
    log_info: LogInfo = {
        "logFileName": file_name
    }

    res: TestResponse = client.post(api, json=log_info)

    if res.status_code == 200:
        raise AssertionError(f"Expected adding the log to fail: {res.data}")
    
    content: dict[str, Any] = json.loads(res.data)
    msg: str = content["content"]
    status: str = content["status"]
    expected_msg: str = "missing key in response"

    assert expected_msg == msg.lower() and status == "error"

def test_create_zip(tmp_path: Path, client: FlaskClient):
    response: TestResponse = client.get(f"/api/packages/{ZIP_NAME}")

    if response.status_code != 200:
        raise AssertionError(f"Got {response.status_code}: {response.data.decode()}")
    
    zip_path: Path = Path(tmp_path / "build" / ZIP_NAME)

    assert zip_path.exists()

def test_create_zip_race(tmp_path: Path, client: FlaskClient):
    api: str = f"/api/packages/{ZIP_NAME}"
    responses: list[TestResponse] = []

    def get() -> None:
        response: TestResponse = client.get(api)
        responses.append(response)

    content: str = ""
    status: str = ""
    race_status: bool = False
    for _ in range(3): 
        if race_status:
            break

        threads: list[Thread] = [Thread(target=get) for _ in range(2)]
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        for res in responses:
            if res.status_code != 200:
                d: dict[str, Any] = json.loads(res.data)
                content = d["content"]
                status = d["status"]
                race_status = True
    
    if not race_status:
        raise AssertionError(f"Max attempts reached for race condition, failed to test: {race_status}")
        
    zip_path: Path = tmp_path / "build" / ZIP_NAME
    if not zip_path.exists():
        raise AssertionError(f"Got {zip_path.name} while checking for race conditions on: {zip_path.parent}")

    assert "updated" in content and status == "error"

def test_get_zip(tmp_path: Path, client: FlaskClient):
    # create the zip first before accessing the api
    log: Log = ttils.get_log(tmp_path)
    zipper: Zip = Zip(tmp_path / ZIP_NAME, log)
    res: dict[str, Any] = zipper.start_zip(tmp_path / "dist")

    zip_size: int = res["files"]["size"]

    response: TestResponse = client.get(f"/api/packages/{ZIP_NAME}")
    if response.status_code != 200:
        raise AssertionError(f"Got {response.status_code}: {response.data.decode()}")

    test_zip_name: str = "out-deploy.zip"
    with open(tmp_path / test_zip_name, "wb") as file:
        file.write(response.data)
    if not (tmp_path / test_zip_name).exists():
        raise AssertionError(f"{test_zip_name} failed to get created")

    zip_file: ZipFile = ZipFile(tmp_path / test_zip_name) 

    assert len(zip_file.filelist) == zip_size