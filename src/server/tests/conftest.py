from flask import Flask
from flask.testing import FlaskClient
from pathlib import Path
from app import create_app
from app_types import Config
from configuration import ZIP_NAME
from . import t_utils as ttils
import pytest

@pytest.fixture()
def app_(tmp_path: Path):
    test_config: Config = {
        "keys_path": tmp_path / "keys",
        "log_path": tmp_path / "logs",
        "log_server_path": tmp_path / "logs" / "server",
        "log_levels": {"stream_level": 10},
        "zip_path": tmp_path / "build" / ZIP_NAME,
        "dist_path": tmp_path / "dist",
        "testing": True
    }

    # creating the necessary files in the dist folder
    ttils.setup(test_config["dist_path"], files=[ttils.BIN_ARGS["arm"], ttils.BIN_ARGS["x86_64"]]) 

    app: Flask = create_app(test_config)

    yield app

@pytest.fixture()
def client(app_: Flask) -> FlaskClient:
    return app_.test_client()