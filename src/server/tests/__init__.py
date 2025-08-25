from pathlib import Path
import sys, os

# fucking python's stupid ass imports. i see why people dont like this langauge
# the working directory is server/
server_path: Path = Path(__file__).parent.parent
sys.path.insert(0, str(server_path))
if os.getcwd() != sys.path[0]:
    os.chdir(sys.path[0])