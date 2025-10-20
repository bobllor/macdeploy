from flask import Blueprint, request, jsonify
from logger import Log
from system.zipper import Zip
from typing import Any
from app_types import Config
import secrets
import system.utils as utils

class ZipUpdater():
    def __init__(self, *, config: Config, logger: Log):
        '''ZIP updating class blueprint.
        
        Parameters
        ----------
            config: Config
                A dictionary of configuration settings.
        '''
        self.logger: Log = logger
        self.config: Config = config

    def get_blueprint(self) -> Blueprint:
        bp: Blueprint = Blueprint("zip_updater", __name__)

        @bp.route("/api/zip/update", methods=["GET"])
        def update_zip():
            '''Updates the ZIP file. A token is used to authenticate the request, and 
            the stored token will be regenerated.

            This endpoint should only be accessed be some type of scheduler.
            '''
            self.logger.debug(f"ZIP updater API accessed by {request.remote_addr}")
            h_token: str = request.headers.get("x-zip-token")
            
            # used to ensure during runtime the file exists.
            if not self.config["token_path"].exists():
                utils.write_to_file(self.config["token_path"], secret_token)

            secret_token: str = utils.read_from(self.config["token_path"])

            if h_token != secret_token:
                self.logger.info("Unauthorized access: %s", h_token)
                return jsonify(utils.generate_response(status="error", content="Unauthorized access")), 401

            zipper: Zip = Zip(self.config["zip_path"], self.logger)
            zip_response: dict[str, Any] = zipper.start_zip()

            if zip_response["status"] == "error":
                return jsonify(zip_response), 500

            self.logger.info("ZIP updated access, token refreshed")

            secret_token = secrets.token_hex(self.config["token_bits"])
            utils.write_to_file(self.config["token_path"], secret_token)

            return jsonify(zip_response), 200
        
        return bp