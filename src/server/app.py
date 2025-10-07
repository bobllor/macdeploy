from flask import Flask, send_file, request, jsonify
from system.process import Process
from system.vars import Vars
from pathlib import Path
from logger import Log
from logging import INFO
from concurrent.futures import ThreadPoolExecutor, Future
from system.zipper import Zip
from typing import Any
import system.system_types as types
import system.utils as utils
import threading
import secrets, os

app: Flask = Flask(__name__)
logger: Log = Log(__name__, levels={"stream_level": INFO})
process: Process = Process(log_dir=Vars.SERVER_LOG_PATH.value, log=logger)

TOKEN_BITS: int = 32
secret_token: str = secrets.token_hex(TOKEN_BITS)

token_file_path: str = f"{Vars.SERVER_PATH.value}/.token"
utils.write_to_file(token_file_path, secret_token)

# change the working directory to the project root folder
curr_path: str = os.getcwd()
if curr_path != Vars.ROOT_PATH.value:
    os.chdir(Vars.ROOT_PATH.value)

@app.route("/")
def home():
    return "Nothing to see here!"

@app.route(f"/api/packages/{Vars.ZIP_FILE_NAME.value}", methods=["GET"])
def get_client_files():
    '''Returns a ZIP file for the client to begin deployment.
    
    The API is strictly used for serving the file. A scheduler to zip the files
    is required to ensure a zip file exists and is updated.
    '''
    zip_file_path: str = f"{Vars.ROOT_PATH.value}/{Vars.ZIP_FILE_NAME.value}"

    # second check after init during runtime. 
    zip_path_obj: Path = Path(Vars.DIST_PATH.value)
    pkg_path_obj: Path = Path(Vars.PKG_PATH.value)

    utils.mk_paths([zip_path_obj, pkg_path_obj])

    if not zip_path_obj.exists():
        logger.critical("Unable to find ZIP file %s", zip_file_path)

        return "No ZIP file found", 400

    logger.info("ZIP file accessed")

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

    return jsonify(utils.generate_response(content=data)), 200

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
    # FIXME: proper exception handling here, it's 3 AM bro...
    threading.Thread(target=process.add_log, args=(content,)).start()

    return jsonify({"status": "success", "content": "Added logs to the server"}), 200

@app.route("/api/zip/update", methods=["GET"])
def update_zip():
    '''Updates the ZIP file. A token is used to authenticate the request, and 
    the stored token will be regenerated.

    This endpoint should only be accessed be some type of scheduler.
    '''
    h_token: str = request.headers.get("x-zip-token")
    
    token_path: Path = Path(token_file_path)
    # used to ensure during runtime the file exists.
    if not token_path.exists():
        utils.write_to_file(token_file_path, secret_token)

    secret_token: str = utils.read_from(token_file_path)

    if h_token != secret_token:
        logger.info("Unauthorized access: %s", h_token)
        return jsonify(utils.generate_response(status="error", content="Unauthorized access")), 401

    zip_path: Path = Path(Vars.ROOT_PATH.value) / Vars.ZIP_FILE_NAME.value

    zipper: Zip = Zip(zip_path, logger)
    zip_response: dict[str, Any] = zipper.start_zip()

    if zip_response["status"] == "error":
        return jsonify(zip_response), 500

    logger.info("ZIP updated access, token refreshed")

    secret_token = secrets.token_hex(TOKEN_BITS)
    utils.write_to_file(token_file_path, secret_token)

    return jsonify(zip_response), 200

if __name__ == '__main__':
    host: str = "0.0.0.0"
    app.run(host=host, debug=True)