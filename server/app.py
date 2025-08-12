from flask import Flask, send_file, request, jsonify
from system.process import Process
from system.vars import Vars
from pathlib import Path
from logger import logger
from concurrent.futures import ThreadPoolExecutor, Future
import system.system_types as types
import system.utils as utils
import threading

app: Flask = Flask(__name__)
process: Process = Process()

@app.route("/")
def home():
    return "Nothing to see here!"

@app.route(f"/api/packages/{Vars.ZIP_FILE_NAME.value}", methods=["GET"])
def get_client_files():
    '''Returns a ZIP file for the client to begin deployment.
    
    The API is strictly used for serving the file. A scheduler to zip the files
    is required to ensure a zip file exists and is updated.
    '''
    zip_file_path: str = f"{Vars.ZIP_PATH.value}/{Vars.ZIP_FILE_NAME.value}"

    # second check after init during runtime. 
    zip_path_obj: Path = Path(Vars.ZIP_PATH.value)
    pkg_path_obj: Path = Path(Vars.PKG_PATH.value)

    utils.mk_paths([zip_path_obj, pkg_path_obj])

    if not zip_path_obj.exists():
        logger.critical("Unable to find ZIP file %s", zip_file_path)

        return "No ZIP file found", 400

    return send_file(zip_file_path)

@app.route("/api/fv", methods=["POST"])
def add_filevault_key():
    '''Adds the FileVault key and the serial tag to the server.
    
    The endpoint only takes POST requests with JSON as the body.
    '''
    content: dict[str, str] = request.get_json()

    logger.debug(f"POST: {content}")

    if not all([key in content for key in ["key", "serialTag"]]):
        logger.warning(f"Invalid POST: {content}")
        return 'Missing expected JSON values "key" or "serial"', 400

    # TODO: figure out how to log this in the same file without needing to make a new log.
    file_vault_key: str = content.get("key")
    serial_tag: str = content.get("serialTag")

    with ThreadPoolExecutor(max_workers=2) as executor:
        future: Future = executor.submit(process.add_filevault, serial_tag, file_vault_key)

        data: str = future.result()

    return jsonify({"status": "success", "content": data}), 200

@app.route("/api/log", methods=["POST"])
def add_log():
    '''Adds the logs from the client device to the server.'''
    content: types.LogInfo = request.get_json()

    logger.debug(f"POST: {content}")

    if not all([key in content for key in ["body", "logFileName"]]):
        logger.warning(f"Invalid POST: {content}")
        return jsonify({"status": "error", "content": 'Missing exepected JSON values "body" or "logFileName"'}), 400
    
    logs_dir_path: Path = Path(Vars.LOGS_PATH.value)
    if not logs_dir_path.exists():
        logs_dir_path.mkdir()

    # NOTE: if the log file is large this could be an issue. maybe look into this in the future? 
    # for now it isn't an issue and probably will not be unless it scales to a large amount...
    threading.Thread(target=process.add_log, args=(content,)).start()

    return jsonify({"status": "success", "content": "Added logs to the server"}), 200

if __name__ == '__main__':
    host: str = "0.0.0.0"
    app.run(host=host)