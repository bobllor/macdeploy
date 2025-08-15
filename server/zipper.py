from system.vars import Vars
from pathlib import Path
from system.utils import get_dir_list
from logger import logger
import subprocess, zipfile, os

def update_zip(zip_path: Path, pkg_path: Path) -> None:
    '''Creates or updates a ZIP file of a package directory 
    using the `zip` command on Unix-like OSes.

    If the ZIP file does not exist, then a new one will be created.

    Parameters
    ----------
        zip_path: Path
            The Path to the ZIP file.
        
        pkg_path: Path
            The directory that contains the packages to install.
    '''
    stripped_pkg_path: str = Vars.PKG_PATH.value.replace(str(pkg_path.parent), "")
    # slice the string to remove the leading slash
    files_to_zip: str = f"{stripped_pkg_path} {Vars.BINARY_NAME.value} {Vars.YAML_CONFIG.value}"[1:]
    new_pkg_path: Path = Path(stripped_pkg_path)

    if not zip_path.exists():
        zip_file: zipfile.ZipFile = zipfile.ZipFile(zip_path, "a")
        
        for path, _, file_list in new_pkg_path.walk():
            for file in file_list:
                path_of_pkg: Path = Path(f"{path}/{file}")

                if path_of_pkg.exists():
                    zip_file.write(path_of_pkg)

        return
    
    zip_file: zipfile.ZipFile = zipfile.ZipFile(zip_path)
    zip_pkg_files: set[str] = set()

    for file in zip_file.filelist:
        if not file.is_dir():
            zip_pkg_files.add(file.filename.lower())

    logger.debug(f"Zip file contents: {zip_pkg_files}")

    # checks for missing packages in the zip file, and updates them with the
    # packages in the deployment files.    
    missing_pkgs: list[str] = []
    server_files: list[str] = get_dir_list(pkg_path, replace_home=True)
    for file in server_files:
        if not file.lower() in zip_pkg_files:
            # removes the absolute path to make it relative.
            file_path: str = f".{Vars.PKG_PATH.value.replace(str(Vars._MAIN_PATH.value), '')}"
            file_name: str = file.replace(Vars._PKG_DIR_NAME.value + "/", "")
        
            missing_pkgs.append(f"{file_path}/{file_name}")

    # missing_pkgs turned out to be useless. keeping it just for logging though.
    # update any missing packages. this will be done subprocess because i am lazy.
    if len(missing_pkgs) > 0:
        logger.info(f"Missing packages in ZIP file: {missing_pkgs}") 

        update_cmd: list[str] = f"zip -ru {str(zip_path)} {files_to_zip}".split()
        execute(update_cmd)

        logger.info(f"Updated ZIP file with files {missing_pkgs}")

def execute(cmd: list[str]) -> None:
    '''Runs a subprocess command for execution for ZIP files.
    
    This is blocking by default, run with threading if non-blocking is required.
    '''
    logger.debug(f"Running command {' '.join(cmd)}")
    try:
        output: subprocess.CompletedProcess = subprocess.run(cmd, capture_output=True)

        stdout: str = output.stdout.decode().strip()
        stderr: str = output.stderr.decode().strip()

        if stdout != "":
            logger.info(stdout)
        if stderr != "":
            logger.info(stderr)
    except Exception as e:
        logger.critical(e)
        # probably not an issue since this is a separate script
        exit()

# NOTE: run this in some type of task scheduler, this is not part of the actual server infrastructure.

# ensures we are in the correct directory.
main_path: Path = Vars._MAIN_PATH.value
curr_path: str = os.getcwd()

logger.debug(f"{__file__.split('/')[-1]} ran in {curr_path}")

if curr_path != str(main_path):
    os.chdir(main_path)
    logger.debug(f"Changed {__file__.split('/')[-1]} to path {str(main_path)}")

pkg_path: Path = Path(Vars.PKG_PATH.value)
zip_path: Path = Path(Vars.ZIP_PATH.value) / Vars.ZIP_FILE_NAME.value

if not zip_path.parent.exists():
    zip_path.parent.mkdir()

update_zip(zip_path, pkg_path)