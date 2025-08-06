from flask import Flask, send_file, Response, request
from system.process import Process
from system.vars import Vars
from pathlib import Path
from system.types import Flags
import system.utils as utils
import threading

# NOTE: after some testing around, i find the best choice is just to work with relative paths.
# too many scenarios can break apart the absolute i built! noted!

# FIXME: change all print statements to logging

app: Flask = Flask(__name__)
process: Process = Process(Path(Vars.PKG_PATH.value)) # FIXME: add the actual pkg location later!

@app.route("/")
def home():
    return "Nothing to see here!"

@app.route("/api/packages/zip", methods=["GET"])
def get_client_files():
    '''Returns a ZIP file for the client to begin deployment.'''
    # second check after init during runtime. 
    zip_path_obj: Path = Path(Vars.ZIP_PATH.value)
    pkg_path_obj: Path = Path(Vars.PKG_PATH.value)
    pkg_hash_path_obj: Path = Path(f"{Vars.SERVER_PATH.value}/{Vars.PKG_HASH_FILE.value}")

    utils.mk_paths([zip_path_obj, pkg_path_obj])
    utils.mk_paths([pkg_hash_path_obj], mk_dir=False)

    # TODO: test if you can still access a ZIP file while a new file is being added to the archive
    can_get_zip: bool = process.flags["zip_status"]
    if not can_get_zip:
        return "ZIP file download has been halted temporarily", 400
    
    cached_hash: str = utils.get_file_content(pkg_hash_path_obj)
    # will always be up-to-date.
    server_hash: str = process.generate_hash(pkg_path_obj)

    print(f"{cached_hash}, {server_hash}")
    # executes the zip file process, this handles if the file doesn't exist
    # and if the cached hash does not match the current server hash.
    if cached_hash == "" or cached_hash != server_hash:
        new_hash: str = process.generate_hash(pkg_path_obj)

        with open(pkg_hash_path_obj, "w") as file:
            file.write(new_hash)

            process.update_zip(zip_path_obj / Vars.ZIP_FILE_NAME.value)

            return "Creating new file", 200

    zip_file_path: str = f"{Vars.ZIP_PATH.value}/{Vars.ZIP_FILE_NAME.value}"
    # send_file(zip_file_path)
    return "wip", 200

@app.route("/api/fv", methods=["POST"])
def add_filevault_key():
    '''Adds the FileVault key and the serial tag to the server.
    
    The endpoint only takes POST requests with JSON as the body.
    '''
    content: dict[str, str] = request.get_json()

    print(f"POST body: {content}")

    if not all([key in content for key in ["key", "serial"]]):
        return 'Missing expected JSON values "key" or "serial"', 400

    file_vault_key: str = content.get("key")
    serial_tag: str = content.get("serial")

    threading.Thread(target=process.add_filevault, args=(serial_tag, file_vault_key,)).start()

    return "Success", 200

@app.route("/api/log", methods=["POST"])
def add_log():
    '''Adds the logs from the client device to the server.'''
    content: dict[str, str] = request.get_json()

    if not all([key in content for key in ["body", "logFileName"]]):
        return 'Missing exepected JSON values "body" or "logFileName"', 400

    # NOTE: if the log file is large this could be an issue. maybe look into this in the future? 
    log_body: str = content["body"]
    file_name: str = content["logFileName"]

    return "wip", 200

if __name__ == '__main__':
    host: str = "0.0.0.0"
    app.run(host=host, debug=True)