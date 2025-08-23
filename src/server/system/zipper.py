from .vars import Vars
from pathlib import Path
from .utils import get_dir_list
from logger import logger
import subprocess, zipfile, os

class Zip:
    def __init__(self, zip_path: Path):
        self.zip_path: Path = zip_path

    def start_zip(self, pkg_path: Path) -> None:
        '''Starts the zipping process for the ZIP file.
        
        This will create a new ZIP file if missing or update an existing ZIP file.

        Parameters
        ----------
            pkg_path: Path
                The directory that contains the packages to install.
        '''
        # ensures we are in the correct directory.
        main_path: Path = Vars.ZIP_PATH.value
        curr_path: str = os.getcwd()

        logger.debug(f"{__file__.split('/')[-1]} ran in {curr_path}")

        if curr_path != str(main_path):
            os.chdir(main_path)
            logger.debug(f"Changed {__file__.split('/')[-1]} to path {str(main_path)}")

        if not self.zip_path.parent.exists():
            self.zip_path.parent.mkdir()

        self._update_zip(pkg_path)

    def _update_zip(self, pkg_path: Path) -> None:
        '''Creates or updates a ZIP file of a package directory 
        using the `zip` command on Unix-like OSes.

        If the ZIP file does not exist, then a new one will be created.

        Parameters
        ----------
            pkg_path: Path
                The directory that contains the packages to install.
        '''
        # ignores the leading slash
        files_to_zip: str = f"{pkg_path.name} {Vars.BINARY_NAME.value}"

        if not self.zip_path.exists():
            zip_file_obj: zipfile.ZipFile = zipfile.ZipFile(self.zip_path, "a")
            
            for path, _, file_list in pkg_path.walk():
                for file in file_list:
                    path_of_pkg: Path = Path(f"{path.name}/{file}")

                    if path_of_pkg.exists():
                        zip_file_obj.write(path_of_pkg)
                    
            zip_file_obj.write(Vars.BINARY_NAME.value)
            zip_file_obj.close()

            return
        
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
                # removes the absolute path to make it relative.
                file_path: str = f".{Vars.PKG_PATH.value.replace(str(Vars._MAIN_PATH.value), '')}"
                file_name: str = file.replace(Vars._PKG_DIR_NAME.value + "/", "")
            
                missing_pkgs.append(f"{file_path}/{file_name}")

        # missing_pkgs turned out to be useless. keeping it just for logging though.
        # update any missing packages. this will be done subprocess because i am lazy.
        if len(missing_pkgs) > 0:
            logger.info(f"Missing packages in ZIP file: {missing_pkgs}") 

            update_cmd: list[str] = f"zip -ru {str(self.zip_path)} {files_to_zip}".split()
            self._execute(update_cmd)

            logger.info(f"Updated ZIP file with files {missing_pkgs}")

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