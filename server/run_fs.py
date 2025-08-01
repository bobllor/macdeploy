from flask import Flask, send_file, Response, request
from pathlib import Path
import subprocess

app = Flask(__name__)

@app.route("/")
def home():
    return "Nothing to see here!"

@app.route("/api/packages/pkg-zip")
def get_client_files() -> Response:
    '''Returns a ZIP file of the client-files directory.'''

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
    app.run(host="10.173.128.112")