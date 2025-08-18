from pathlib import Path
import sys

server_path: Path = Path(__file__)

# yeah... python stuff i guess.
if sys.path[0] != str(server_path):
    sys.path.insert(0, str(server_path.resolve().parent))