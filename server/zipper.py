from system.vars import Vars
from pathlib import Path
from system.utils import get_dir_list
from logger import logger
import subprocess, zipfile, os

def update_zip(zip_path: Path, pkg_path: Path) -> None:
    '''Updates a ZIP file with missing packages. This compares the current packages
    inside the package directory and the packages in the ZIP file, and updates 
    missing packages to the ZIP file.

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
    if not zip_path.exists():
        # files to zip: packages folder, go binary, config.yaml
        # the paths must be relative, absolute paths introduces issues with zipping.
        zip_cmd: list[str] = f"zip -r {str(zip_path)} {files_to_zip}".split()

        execute(zip_cmd)
        logger.info(f"ZIP file created in {Vars.ZIP_PATH.value}")

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
            file_path: str = f".{Vars.PKG_PATH.value.replace(str(Vars._MAIN_PATH.value), "")}"
            file_name: str = file.replace(Vars._PKG_DIR_NAME.value + "/", "")
        
            missing_pkgs.append(f"{file_path}/{file_name}")

    # missing_pkgs turned out to be useless. keeping it just for logging though.
    if len(missing_pkgs) > 0:
        print(f"Missing packages in ZIP file: {missing_pkgs}") 

        update_cmd: list[str] = f"zip -ru {str(zip_path)} {files_to_zip}".split()
        execute(update_cmd)

        print(f"Updated ZIP file with files {missing_pkgs}")
    else:
        print("No updates needed for the ZIP file")

def execute(cmd: list[str]) -> None:
    '''Runs a subprocess command for execution for ZIP files.
    
    This is blocking by default, run with threading if non-blocking is required.
    '''
    print(f"Running command {" ".join(cmd)}")
    output: subprocess.CompletedProcess = subprocess.run(cmd, capture_output=True)

    err: str = str(output.stderr)

    print("Command finished")

# NOTE: run this in some type of task scheduler, this is not called in the actual server code.

# ensures we are in the correct directory.
main_path: Path = Vars._MAIN_PATH.value
curr_path: str = os.getcwd()

print(f"{__file__.split("/")[-1]} ran in {curr_path}")

if curr_path != str(main_path):
    os.chdir(main_path)
    print(f"Changed {__file__.split("/")[-1]} to path {str(main_path)}")

pkg_path: Path = Path(Vars.PKG_PATH.value)
zip_path: Path = Path(Vars.ZIP_PATH.value) / Vars.ZIP_FILE_NAME.value
update_zip(zip_path, pkg_path)