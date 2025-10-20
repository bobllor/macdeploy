from configuration import ARM_BINARY_NAME, X86_BINARY_NAME, DIST_PATH
from pathlib import Path
from .utils import get_dir_list, generate_response
from logger import Log
from typing import TypedDict, Any
from contextlib import contextmanager
from collections.abc import Callable
import zipfile, os

class BinaryArgs(TypedDict):
    arm: str
    x86_64: str

class Zip:
    def __init__(self, zip_path: Path, log: Log, *, binary_args: BinaryArgs = {}):
        '''Create and update ZIP files.
        
        Parameters
        ----------
            zip_path: Path
                The Path object of the path to the ZIP file.
            
            log: Log
                The logger Log of the program.
            
            binary_args: BinaryArgs, default {}
                Dictionary arguments replacing the two binary names to a given name.
                Only used for testing.
        '''
        self.zip_path: Path = zip_path
        self.log: Log = log

        self.arm_binary: str = binary_args.get("arm", ARM_BINARY_NAME)
        self.x86_binary: str = binary_args.get("x86_64", X86_BINARY_NAME)

        # cache for recursive parent creation when appending new files in the ZIP
        self._created_files: set[str] = {".", "./", ""}

    def start_zip(self, dist_path: Path = DIST_PATH) -> dict[str, Any]:
        '''Starts the zipping process for the ZIP file.
        This will create a new ZIP file or update the existing ZIP file.
        
        Upon completion it will return a dictionary for the response.

        Parameters
        ----------
            dist_dir: Path, default Path(path/of/dist)
                The string path to the directory of the dist directory, containing all the files for distribution.
        '''
        dist_path_str: str = str(dist_path)

        if not dist_path.exists():
            dist_path.mkdir()
        
        binary_names: list[str] = [self.arm_binary, self.x86_binary]
        for binary in binary_names:
            if not (dist_path / binary).exists():
                self.log.critical("Binary %s not found in %s", binary, dist_path_str)
                return generate_response(
                    status="error", 
                    content=f"Binary not found on server",
                    files={"size": 0, "content": []},
                    statusCode=500
                )

        try:
            func: Callable[[str], dict[str, Any]] = self._create_zip
            if self.zip_path.exists():
                func = self._update_zip

            with self._zipper(dist_path, func) as res:
                zip_response: dict[str, Any] = res
        except Exception as e:
            # i have no idea what exceptions can happen here.
            # leaving a all-purpose catch, will change over time.
            self.log.critical(f"Failed ZIP process {e}")

            return generate_response(
                status="error", 
                content=f"An unexpected error occurred on the server",
                files={"size": 0, "content": []},
                statusCode=500
            )

        return zip_response
    
    @contextmanager
    def _zipper(self, dist_path: Path, func: Callable[[str], dict[str, Any]]):
        # ensure we are working in the parent of the distribution folder
        cwd: str = os.getcwd()
        parent: str = str(dist_path.parent)
        if cwd != parent:
            self.log.warning("Current directory %s was not in distribution's parent %s", cwd, parent)
            os.chdir(parent)
            self.log.info("Updated working directory to %s", parent)

        res: dict[str, Any] = func(dist_path)
        try:
            yield res
        finally:
            # initially had below, but python should be in the root directory anyways.
            # will keep it just in case.
            pass
            #os.chdir(cwd)
            #self.log.info("Updated working directory to %s", cwd)

    def _create_zip(self, dist_path: Path) -> dict[str, Any]:
        '''Creates the ZIP file of a package directory.
        
        Parameters
        ----------
            dist_path: Path
                The Path object of the dist folder.
        '''
        zip_file_obj: zipfile.ZipFile = zipfile.ZipFile(self.zip_path, "a")
        self.log.warning("ZIP file does not exist")
        self.log.debug("Searching in path %s, working directory: %s", str(dist_path), os.getcwd())

        zip_contents: list[str] = []

        root_path: str = str(dist_path.parent)

        for path, _, file_list in dist_path.walk():
            # required to make the directory paths relative and not absolute.
            replaced_parent: Path = Path(str(path).replace(root_path + "/", ""))
            # this creates the "folder" inside the zip, which mimics the way zip works on unix.
            self._recurse_create_directories(replaced_parent, zip_file_obj, self._created_files)
            for file in file_list:
                relative_file_path: Path = replaced_parent / file

                if relative_file_path.exists():
                    if not relative_file_path.is_dir():
                        zip_file_obj.write(relative_file_path)

                        self.log.info("File %s added", str(relative_file_path))
                else:
                    self.log.error("Issue searching path %s in pwd: %s", str(relative_file_path), os.getcwd())

        zip_len: int = len(zip_file_obj.filelist)
        
        for file in zip_file_obj.filelist:
            zip_contents.append(file.filename)

        zip_file_obj.close()

        self.log.info("Created ZIP file at %s", str(self.zip_path.absolute()))
        self.log.debug(f"New ZIP: {zip_contents}")
        self.log.debug(f"New ZIP length: {zip_len}")

        return generate_response(
            status="success", 
            content="ZIP file created",
            files={"size": zip_len, "content": zip_contents},
            statusCode=200
        )

    def _update_zip(self, dist_path: Path) -> dict[str, Any]:
        '''Updates the ZIP file of using the `zip` command on Unix-like OSes.

        A dictionary response is returned upon success or failure.

        Parameters
        ----------
            dist_path: Path
                The Path object of the dist folder.
        '''
        zip_file_obj: zipfile.ZipFile = zipfile.ZipFile(self.zip_path, mode="a")

        root_path: str = str(dist_path.parent)
        for file in zip_file_obj.filelist:
            updated_file: str = file.filename.lower().replace(root_path + "/", "")
            
            if file.is_dir():
                # drops the ending slash, ZipFile appends it to directories.
                updated_file = updated_file[:-1]

            self._created_files.add(updated_file)

        self.log.debug(f"Existing files: {self._created_files}")

        # checks the files in the current server directory to the files in the
        # ZIP. this is to get the files that do not exist.
        new_files: list[str] = []
        server_files: list[str] = get_dir_list(dist_path, replace_root=True)
        for file in server_files:
            updated_file: str = file.lower().replace(root_path + "/", "")

            if not updated_file in self._created_files:
                # drops the full path up to the parent from the file name
                new_files.append(updated_file)
                self.log.debug("Found new file %s", updated_file)

        content_msg: str = "No new files found to update ZIP"
        added_count: int = len(new_files)
        original_count: int = len(zip_file_obj.filelist)

        # new_files turned out to be useless. keeping it just for logging though.
        # update any missing packages. this will be done subprocess because i am lazy.
        if added_count > 0:
            self.log.debug(f"Missing packages in ZIP file: {new_files}") 

            # files are already trimmed to relative paths with the loop above
            for file in new_files:
                file_path: Path = Path(file)

                if file_path.exists():
                    self._recurse_create_directories(file_path, zip_file_obj, self._created_files)

                    if not file_path.is_dir():
                        zip_file_obj.write(file_path)

                        self.log.info("File %s added", str(file_path))

            self.log.info(f"Updated ZIP file with files {new_files}")
            self.log.debug(
                f"Original ZIP length: {original_count} | New length: {added_count + original_count}"
            )

            zip_file_obj.close()

            content_msg = f"ZIP file updated with {added_count} {"file" if added_count == 1 else "files"}"
        else:
            self.log.info(f"No new files found, skipping ZIP update")

        return generate_response(
            status="success", 
            content=content_msg,
            files={"size": len(new_files), "content": new_files},
            statusCode=200
        )
    
    def _recurse_create_directories(self, path: Path, zip_obj: zipfile.ZipFile, created_paths: set[str]) -> None:
        '''Writes to the ZIP file by recursively going through each parent until
        the base case is reached.

        This is to simulate the zip -ru command, which creates the files and its parent
        folders.
        '''
        parent: str = str(path.parent)
        if parent in created_paths and str(path) in created_paths:
            return

        if path.is_dir():
            zip_obj.write(path)
            self.log.info("Directory %s added", path)

        created_paths.add(str(path))
        created_paths.add(parent)
        self._recurse_create_directories(path.parent, zip_obj, created_paths)