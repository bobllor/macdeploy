from flask import Flask
import configuration as conf
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
if curr_path != str(conf.ROOT_PATH):
    os.chdir(conf.ROOT_PATH)

def create_app(config_arg: Config = None) -> Flask:
    app: Flask = Flask(__name__)
    config: Config = {
        "log_levels": {"stream_level": INFO, "log_level": DEBUG},
        "zip_path": conf.ZIP_PATH / conf.ZIP_NAME,
        "log_path": conf.LOGS_PATH,
        "log_server_path": conf.SERVER_LOGS_PATH,
        "dist_path": conf.DIST_PATH,
        "keys_path": conf.KEYS_PATH,
        "testing": False,
        "token_path": conf.SERVER_PATH / ".token",
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

    # i love unit tests! unit tests make logging so fun!!
    updater: ZipUpdater = ZipUpdater(config=app.config, logger=logger)
    processor: Processor = Processor(config=app.config, logger=logger)
    requestors: Requestors = Requestors(config=app.config, logger=logger)

    app.register_blueprint(updater.get_blueprint())
    app.register_blueprint(processor.get_blueprint())
    app.register_blueprint(requestors.get_blueprint())

    return app

if __name__ == '__main__':
    host: str = "0.0.0.0"
    app: Flask = create_app()

    app.run(host=host, debug=True)