from flask import Flask
from logger import Log
from logging import INFO, DEBUG, setLoggerClass
from app_types import Config
from blueprints.processor import Processor
from blueprints.zip_updater import ZipUpdater
from blueprints.requestors import Requestors
from blueprints.query import Query
from pathlib import Path
import system.utils as utils
import configuration as conf
import secrets
import os

# change the working directory to the project root folder
curr_path: str = os.getcwd()
if curr_path != str(conf.ROOT_PATH):
    os.chdir(conf.ROOT_PATH)

class EnvValues:
    '''Holds environmental variable values.'''
    def __init__(self, *, 
        root_dir: str, 
        host: str
    ):
        self.HOST: str = host
        self.ROOT_DIR: str = root_dir

def create_app(config_arg: Config = None, *, root: Path | str = None) -> Flask:
    app: Flask = Flask(__name__)
    if root is None:
        root = conf.ROOT_PATH

    # handles strings, path is not affected 
    root = Path(root)

    config: Config = {
        "log_levels": {"stream_level": INFO, "log_level": DEBUG},
        "zip_path": root / conf.ZIP_NAME,
        "log_path": root / conf.LOGS_NAME,
        "log_server_path": root / conf.LOGS_NAME / conf.SERVER_LOGS_NAME ,
        "dist_path": root / conf.DIST_DIR_NAME,
        "keys_path": root / conf.KEYS_NAME,
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

    logger.debug(f"Root directory: {root.absolute()}")

    # i love unit tests! unit tests make logging so fun!!
    updater: ZipUpdater = ZipUpdater(config=app.config, logger=logger)
    processor: Processor = Processor(config=app.config, logger=logger)
    requestors: Requestors = Requestors(config=app.config, logger=logger)
    query: Query = Query(config=app.config, logger=logger)

    app.register_blueprint(updater.get_blueprint())
    app.register_blueprint(processor.get_blueprint())
    app.register_blueprint(requestors.get_blueprint())
    app.register_blueprint(query.get_blueprint())

    return app

def get_env() -> EnvValues:
    '''Retrieves the environmental variables and loads them into the program.
    This does not load external .env files.

    If an ENV key is empty, then it will default to production values.
    '''
    FLASK_HOST: str = "FLASK_HOST"
    FLASK_ROOT_DIR: str = "FLASK_ROOT_DIR"

    host: str = os.getenv(FLASK_HOST, "0.0.0.0")
    root_dir: str = os.getenv(FLASK_ROOT_DIR, str(conf.ROOT_PATH.absolute()))

    env_values: EnvValues = EnvValues(
        host=host, 
        root_dir=root_dir,
    )

    return env_values

if __name__ == '__main__':
    host: str = "0.0.0.0"
    env: EnvValues = get_env()
    
    app: Flask = create_app(root=env.ROOT_DIR)

    app.run(host=env.HOST, debug=True)