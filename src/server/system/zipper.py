from .vars import Vars
from pathlib import Path
from .utils import get_dir_list
from logger import logger
from typing import TypedDict
import subprocess, zipfile, os, pdb

class PathArgs(TypedDict):
    dist_dir: str
    binary_file: str

class Zip:
    def __init__(self, zip_path: Path, *, path_args: PathArgs = {}):
        '''Create and update ZIP files.
        
        Parameters
        ----------
            zip_path: Path
                The Path object of the path to the ZIP file.
            
            path_args: PathArgs, default {}
                Dictionary arguments replacing paths in the class. This should not be
                used except during testing.
        '''
        self.zip_path: Path = zip_path

        self.arm_binary: str = path_args.get("arm_binary", Vars.ARM_BINARY_NAME.value)
        self.x86_binary: str = path_args.get("x86_binary", Vars.X86_BINARY_NAME.value)

    def start_zip(self, dist_path: Path = Path(Vars.DIST_PATH.value)) -> tuple[bool, str]:
        '''Starts the zipping process for the ZIP file.
        This will create a new ZIP file or update the existing ZIP file.
        
        Upon completion it will return a boolean indicating its operational status, and a string for
        additional information.

        Parameters
        ----------
            dist_dir: Path, default Path(Vars.DIST_PATH.value)
                The string path to the directory of the dist directory, containing all the files for distribution.
        '''
        dist_path_str: str = str(dist_path)

        if not dist_path.exists():
            dist_path.mkdir()
        
        binary_names: list[str] = [self.arm_binary, self.x86_binary]
        for binary in binary_names:
            if not (dist_path / binary).exists():
                logger.critical("Binary %s not found in %s", binary, dist_path_str)
                return False, f"Binary not found on server"

        try:
            if not self.zip_path.exists():
                zip_status: str = self._create_zip(dist_path)
            else:
                zip_status: str = self._update_zip(dist_path)
        except Exception as e:
            # i have no idea what exceptions can happen here.
            # leaving a all-purpose catch, will change over time.
            logger.critical(f"Failed ZIP process {e}")

            return False, "An unexpected error occurred on the server"

        return True, zip_status

    def _create_zip(self, dist_path: Path) -> str:
        '''Creates the ZIP file of a package directory.
        
        Parameters
        ----------
            dist_path: Path
                The Path object of the dist folder.
        '''
        zip_file_obj: zipfile.ZipFile = zipfile.ZipFile(self.zip_path, "a")
        logger.warning("ZIP file does not exist")
        logger.debug("Searching in path %s, working directory: %s", str(dist_path), os.getcwd())

        zip_contents: list[str] = []

        for path, _, file_list in dist_path.walk():
            for file in file_list:
                # adds the pkg_path value to nested directories to keep structure
                # this works for both the test and production. did this at 2 am i have a brain aneurysm
                file_name: str = f"{str(path)}/{file}".replace(Vars.ROOT_PATH.value + "/", "")

                path_of_pkg: Path = Path(file_name)

                if path_of_pkg.exists():
                    zip_file_obj.write(path_of_pkg)
                    zip_contents.append(path_of_pkg.name)
                else:
                    logger.error("Issue searching path %s in pwd: %s", str(path_of_pkg), os.getcwd())
                
        zip_file_obj.close()

        logger.info("Created ZIP file at %s", str(self.zip_path.absolute()))
        logger.debug(f"New ZIP contents: {zip_contents}")

        return "Created ZIP file"

    def _update_zip(self, dist_path: Path) -> str:
        '''Updates the ZIP file of using the `zip` command on Unix-like OSes.

        Upon success, a string is returned indicating its status.

        Parameters
        ----------
            dist_path: Path
                The Path object of the dist folder.
        '''
        zip_file_obj: zipfile.ZipFile = zipfile.ZipFile(self.zip_path)
        zip_pkg_files: set[str] = set()

        for file in zip_file_obj.filelist:
            if not file.is_dir():
                zip_pkg_files.add(file.filename.lower())

        logger.debug(f"Zip file contents: {zip_pkg_files}")

        # checks for missing packages in the zip file, and updates them with the
        # packages in the deployment files.    
        missing_pkgs: list[str] = []
        server_files: list[str] = get_dir_list(dist_path, replace_home=True)
        for file in server_files:
            # lol... splitting the dist directory and taking the last files for relative paths
            # and removing the leading slash with a slice
            file_name: str = file.lower()

            if not file_name in zip_pkg_files:
                # drops the full path up to the parent from the file name
            
                missing_pkgs.append(file_name)

        # missing_pkgs turned out to be useless. keeping it just for logging though.
        # update any missing packages. this will be done subprocess because i am lazy.
        if len(missing_pkgs) > 0:
            logger.info(f"Missing packages in ZIP file: {missing_pkgs}") 

            # relative path is needed to skip the full path creation
            # of the zip command while maintaining its folder structure.
            update_cmd: list[str] = f'zip -ru {str(self.zip_path)} {dist_path.name}'.split()
            self._execute(update_cmd)

            logger.info(f"Updated ZIP file with files {missing_pkgs}")

        return "Updated ZIP file"

    def _execute(self, cmd: list[str]) -> None:
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