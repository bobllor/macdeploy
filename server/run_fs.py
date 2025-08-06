from flask import Flask, send_file, Response, request
from system.process import Process
from system.vars import Vars
from pathlib import Path
from system.types import Flags
import system.utils as utils

# NOTE: after some testing around, i find the best choice is just to work with relative paths.
# too many scenarios can break apart the absolute i built! noted!

# FIXME: change all print statements to logging

app: Flask = Flask(__name__)
process: Process = Process(Path(Vars.PKG_PATH.value)) # FIXME: add the actual pkg location later!

flags: Flags = {
    "zipping_status": False
}

@app.route("/")
def home():
    return "Nothing to see here!"

@app.route("/api/packages/zip", methods=["GET"])
def get_client_files():
    '''Returns a ZIP file for the client to begin deployment.'''
    zip_file_path: str = f"{Vars.ZIP_PATH.value}/{Vars.ZIP_FILE_NAME.value}"

    # second check after init during runtime. 
    zip_path_obj: Path = Path(Vars.ZIP_PATH.value)
    pkg_path_obj: Path = Path(Vars.PKG_PATH.value)
    utils.mk_paths([zip_path_obj, pkg_path_obj])
    pkg_hash_path_obj: Path = Path(f"{Vars.SERVER_PATH.value}/{Vars.PKG_HASH_FILE.value}")
    utils.mk_paths([pkg_hash_path_obj], mk_dir=False)

    # TODO: add something here for missing zip file.
    '''if not Path(zip_file_path).exists():
        return "", 400
    '''
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

        should_zip: bool = flags["zipping_status"]
        if not should_zip: 
            flags["zipping_status"] = True

        return "Creating new file, try again in 5 minutes.", 200

    process.update_zip(zip_path_obj / Vars.ZIP_FILE_NAME.value)
    # send_file(zip_file_path)
    return "wip", 200

@app.route("/api/fv", methods=["POST"])
def add_filevault_key():
    '''Adds the FileVault key and the serial tag to the server.
    
    The endpoint only takes POST requests with JSON as the body.
    '''
    content: dict[str, str] = request.get_json()

    # TODO: add logging here

    if not all([key in content for key in ["key", "serial"]]):
        return "", 400

    file_vault_key: str = content.get("key")
    serial_tag: str = content.get("serial")

    process.add_filevault(serial_tag, file_vault_key)

    return "", 200

if __name__ == '__main__':
    host: str = "0.0.0.0"
    app.run(host=host, debug=True)