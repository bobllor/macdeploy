from flask import Flask
from system.vars import Vars
from pathlib import Path
from logger import Log
from logging import INFO, DEBUG, setLoggerClass
from app_types import Config
from blueprints.processor import Processor
from blueprints.zip_updater import ZipUpdater
from blueprints.requestors import Requestors
import system.utils as utils
import secrets, os

# change the working directory to the project root folder
curr_path: str = os.getcwd()
if curr_path != Vars.ROOT_PATH.value:
    os.chdir(Vars.ROOT_PATH.value)

def create_app(config_arg: Config = None) -> Flask:
    app: Flask = Flask(__name__)
    config: Config = {
        "log_levels": {"stream_level": INFO, "log_level": DEBUG},
        "zip_path": Path(Vars.ZIP_PATH.value) / Vars.ZIP_FILE_NAME.value,
        "log_path": Path(Vars.LOGS_PATH.value),
        "log_server_path": Path(Vars.SERVER_LOG_PATH.value),
        "dist_path": Path(Vars.DIST_PATH.value),
        "keys_path": Path(Vars.KEYS_PATH.value),
        "testing": False,
        "token_path": Path(Vars.SERVER_PATH.value) / ".token",
        "token_bits": 32,
    }

    secret_token: str = secrets.token_hex(config["token_bits"])
    utils.write_to_file(config["token_path"], secret_token)

    # replacing the values of default config.
    if config_arg is not None:
        for key, val in config_arg.items():
            config[key] = val

    app.config.update(config)
    logger: Log = Log(log_path=config["log_server_path"], levels=app.config["log_levels"])
    logger.set_logger()

    updater: ZipUpdater = ZipUpdater(config=app.config)
    processor: Processor = Processor(config=app.config)
    requestors: Requestors = Requestors(config=app.config)

    app.register_blueprint(updater.get_blueprint())
    app.register_blueprint(processor.get_blueprint())
    app.register_blueprint(requestors.get_blueprint())

    return app

if __name__ == '__main__':
    host: str = "0.0.0.0"
    app: Flask = create_app()

    app.run(host=host, debug=True)