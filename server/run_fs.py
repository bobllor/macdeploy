from flask import Flask, send_file, Response, request
from system.process import Process
from system.vars import Vars
from pathlib import Path
import system.types as types
import system.utils as utils
import threading

# NOTE: after some testing around, i find the best choice is just to work with relative paths.
# too many scenarios can break apart the absolute i built! noted!

# FIXME: change all print statements to logging

app: Flask = Flask(__name__)
process: Process = Process() # FIXME: add the actual pkg location later!

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

    return send_file(zip_file_path)

@app.route("/api/fv", methods=["POST"])
def add_filevault_key():
    '''Adds the FileVault key and the serial tag to the server.
    
    The endpoint only takes POST requests with JSON as the body.
    '''
    content: dict[str, str] = request.get_json()

    print(f"POST body: {content}")

    if not all([key in content for key in ["key", "serial"]]):
        return 'Missing expected JSON values "key" or "serial"', 400

    # TODO: figure out how to log this in the same file without needing to make a new log.
    file_vault_key: str = content.get("key")
    serial_tag: str = content.get("serial")

    threading.Thread(target=process.add_filevault, args=(serial_tag, file_vault_key,)).start()

    return "Success", 200

@app.route("/api/log", methods=["POST"])
def add_log():
    '''Adds the logs from the client device to the server.'''
    content: types.LogInfo = request.get_json()

    if not all([key in content for key in ["body", "logFileName"]]):
        return 'Missing exepected JSON values "body" or "logFileName"', 400
    
    logs_dir_path: Path = Path(Vars.LOGS_PATH.value)
    if not logs_dir_path.exists():
        logs_dir_path.mkdir()

    # NOTE: if the log file is large this could be an issue. maybe look into this in the future? 
    # for now it isn't an issue and probably will not be unless it scales to a large amount...
    threading.Thread(target=process.add_log, args=(content,)).start()

    return "Success", 200

if __name__ == '__main__':
    host: str = "0.0.0.0"
    app.run(host=host, debug=True)