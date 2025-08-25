from .vars import Vars
from pathlib import Path
from .utils import get_dir_list
from logger import logger
from typing import TypedDict
import subprocess, zipfile, os

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

        self.dist_path: str = path_args.get("dist_dir", Vars.DIST_PATH.value)
        self.binary_file: str = path_args.get("binary_file", Vars.BINARY_NAME.value)

    def start_zip(self, pkg_path: Path, *, dist_dir: str = Vars.DIST_PATH.value) -> tuple[bool, str]:
        '''Starts the zipping process for the ZIP file.
        This will create a new ZIP file or update the existing ZIP file.
        
        Upon completion it will return a boolean indicating its operational status, and a string for
        additional information.

        Parameters
        ----------
            pkg_path: Path
                The directory that contains the packages to install.
            
            dist_dir: str
                The string path to the directory of the dist directory. By default it already points
                to the dist directory, this is used only for testing.
        '''
        # ensures we are in the dist directory during this.
        curr_path: str = os.getcwd()

        logger.debug(f"{__file__.split('/')[-1]} ran in {curr_path}")

        # swapping to dist path because of the zip command.
        # bypassing the full path zip creation, this will work with relative paths.
        if curr_path != dist_dir:
            os.chdir(dist_dir)
            logger.debug(f"Changed {__file__.split('/')[-1]} to path {dist_dir}")

        if not self.zip_path.parent.exists():
            self.zip_path.parent.mkdir()

        if not (Path(self.dist_path) / self.binary_file).exists():
            logger.critical("Binary not found in %s", self.dist_path)
            return False, f"Binary not found on server"

        try:
            if not self.zip_path.exists():
                zip_status: str = self._create_zip(pkg_path)
            else:
                zip_status: str = self._update_zip(pkg_path)
        except Exception as e:
            # i have no idea what exceptions can happen here.
            # leaving a all-purpose catch, will change over time.
            logger.critical(f"Failed ZIP process {e}")

            return False, "An unexpected error occurred on the server"

        os.chdir(curr_path)

        return True, zip_status

    def _create_zip(self, pkg_path: Path) -> str:
        '''Creates the ZIP file of a package directory.'''
        zip_file_obj: zipfile.ZipFile = zipfile.ZipFile(self.zip_path, "a")
        logger.warning("ZIP file does not exist")
        
        for path, _, file_list in pkg_path.walk():
            # adds the pkg_path value to nested directories to keep structure
            parent: str = path.name if pkg_path.name == path.name else f"{pkg_path.name}/{path.name}"

            for file in file_list:
                path_of_pkg: Path = Path(f"{parent}/{file}")

                if path_of_pkg.exists():
                    zip_file_obj.write(path_of_pkg)
                
        zip_file_obj.write(self.binary_file)
        zip_file_obj.close()

        logger.info("Created ZIP file in %s", self.dist_path)

        return "Created ZIP file"

    def _update_zip(self, pkg_path: Path) -> str:
        '''Updates the ZIP file of using the `zip` command on Unix-like OSes.

        Upon success, a string is returned indicating its status.

        Parameters
        ----------
            pkg_path: Path
                The directory that contains the packages to install.
        '''
        pkg_path_str: str = str(pkg_path)
        
        zip_file_obj: zipfile.ZipFile = zipfile.ZipFile(self.zip_path)
        zip_pkg_files: set[str] = set()

        for file in zip_file_obj.filelist:
            if not file.is_dir():
                zip_pkg_files.add(file.filename.lower())

        logger.debug(f"Zip file contents: {zip_pkg_files}")

        # checks for missing packages in the zip file, and updates them with the
        # packages in the deployment files.    
        missing_pkgs: list[str] = []
        server_files: list[str] = get_dir_list(pkg_path, replace_home=True)
        for file in server_files:
            if not file.lower() in zip_pkg_files:
                # drops the full path up to the parent from the file name
                file_name: str = file.replace(pkg_path_str + "/", "")
            
                missing_pkgs.append(file_name)

        # missing_pkgs turned out to be useless. keeping it just for logging though.
        # update any missing packages. this will be done subprocess because i am lazy.
        if len(missing_pkgs) > 0:
            logger.info(f"Missing packages in ZIP file: {missing_pkgs}") 

            # relative path is needed to skip the full path creation
            # of the zip command while maintaining its folder structure.
            files_to_zip: str = f"{pkg_path.name} {self.binary_file}"

            update_cmd: list[str] = f"zip -ru {str(self.zip_path)} {files_to_zip}".split()
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