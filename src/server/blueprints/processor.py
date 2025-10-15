from flask import Blueprint, request, jsonify
from logging import Logger, getLogger
from system.process import Process
from typing import Any
from app_types import Config
from concurrent.futures import ThreadPoolExecutor, Future
import system.system_types as types

class Processor():
    def __init__(self, *, config: Config):
        '''Processing class blueprint.
        
        Parameters
        ----------
            config: Config
                A dictionary of configuration settings.
        '''
        self.logger: Logger = getLogger("Log")
        self.config: Config = config

    def get_blueprint(self) -> Blueprint:
        bp: Blueprint = Blueprint("processor", __name__)

        @bp.route("/api/fv", methods=["POST"])
        def add_filevault_key():
            '''Adds the FileVault key and the serial tag to the server.
            
            The endpoint only takes POST requests with JSON as the body.
            '''
            process: Process = Process(log=self.logger)
            content: types.KeyInfo = request.get_json()
            self.logger.debug(f"Keys API accessed by {request.remote_addr}")

            with ThreadPoolExecutor(max_workers=2) as executor:
                future: Future = executor.submit(process.add_filevault, content, self.config["keys_path"])

                data: str = future.result()

            return data, data["statusCode"]

        @bp.route("/api/log", methods=["POST"])
        def add_log():
            '''Adds the logs from the client device to the server.'''
            process: Process = Process(log=self.logger)
            self.logger.debug(f"Logs API accessed by {request.remote_addr}")

            content: types.LogInfo = request.get_json()

            with ThreadPoolExecutor(max_workers=2) as executor:
                future: Future = executor.submit(process.add_log, content, self.config["log_path"])

                data: dict[str, Any] = future.result()
            
            return jsonify(data), data["statusCode"]
        
        return bp