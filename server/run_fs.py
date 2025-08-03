from flask import Flask, send_file, Response, request
from pathlib import Path
from system.vars import Vars
import subprocess

app = Flask(__name__)

@app.route("/")
def home():
    return "Nothing to see here!"

@app.route("/api/packages/pkg-zip")
def get_client_files() -> Response:
    '''Returns a ZIP file of the client-files directory.'''
    with open(f"{Vars.SERVER_DIR}/{Vars.VERSION_FILE}", "r") as file:
        version: str = file.read()

        print(version)

    # send_file(Vars.YAML_CONFIG)
    return "wip"

@app.route("/api/fv/")
def add_filevault_key():
    '''Adds the FileVault key and the serial tag to the server.'''
    content: dict[str, str] = request.args.to_dict()

    if not all([key in content for key in ["key", "serial"]]):
        return "Error here..."

    file_vault_key: str = content.get("key")
    serial_tag: str = content.get("serial")

    return "Hello"

if __name__ == '__main__':
    host: str = "127.0.0.1"
    app.run(host=host)