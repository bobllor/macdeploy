from pathlib import Path
import sys

server_path: Path = Path(__file__)

if sys.path[0] != str(server_path):
    sys.path.insert(0, str(server_path.resolve().parent))