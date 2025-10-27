from flask import Blueprint, request, jsonify, send_file
from configuration import ZIP_NAME
from pathlib import Path
from logger import Log
from logging import Logger
from concurrent.futures import ThreadPoolExecutor, Future
from multiprocessing import Lock as Lock_
from multiprocessing.synchronize import Lock
from system.zipper import Zip
from app_types import Config
import system.utils as utils

class Requestors():
    def __init__(self, *, config: Config, logger: Log):
        '''Default GET request routes blueprint.
        
        Parameters
        ----------
            config: Config
                A dictionary of configuration settings.
        '''
        self.logger: Logger = logger
        self.config: Config = config

        self.lock: Lock = Lock_()

    def get_blueprint(self) -> Blueprint:
        bp: Blueprint = Blueprint("requestors", __name__)

        @bp.route("/")
        def home():
            return jsonify(
                utils.generate_response(
                    content="Use the endpoints to start the deployment."
                )), 200

        @bp.route(f"/api/packages/{str(ZIP_NAME)}", methods=["GET"])
        def get_client_files():
            '''Returns a ZIP file for the client to begin deployment.
            
            The API is strictly used for serving the file. A scheduler to zip the files
            is required to ensure a zip file exists and is updated.
            '''
            self.logger.debug(f"ZIP file API accessed by {request.remote_addr}")
            zip_file_path: Path = self.config["zip_path"]

            if not zip_file_path.exists():
                self.logger.warning("Unable to find ZIP file in %s", zip_file_path)

                lock_: Lock = self.lock.acquire(block=False)
                self.logger.debug("Lock status: %r", lock_)
                if lock_:
                    try:    
                        # second check after init during runtime. 
                        utils.mk_paths([zip_file_path.parent])

                        with ThreadPoolExecutor(max_workers=2) as executor:
                            zipper: Zip = Zip(self.config["zip_path"], self.logger)
                            future: Future = executor.submit(zipper.start_zip, self.config["dist_path"])

                            data: str = future.result()

                            if data["status"] == "error":
                                return data, data["statusCode"]
                    finally:
                        self.lock.release()
                else:
                    self.logger.warning(f"ZIP file requested while updating")
                    return utils.generate_response(
                        status="error",
                        content="ZIP file is being updated"
                    ), 500

                return send_file(zip_file_path), 200

            self.logger.info(f"ZIP file successfully requested")

            return send_file(zip_file_path), 200
        
        return bp