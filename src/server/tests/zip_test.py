from pathlib import Path
import subprocess, zipfile, sys

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from system.vars import Vars
from system.utils import unlink_children, get_dir_list

zip_dir: str = Vars.DIST_PATH.value + "-temp"
pkg_dir: str = Vars.PKG_PATH.value + "-temp"

def mk_temp_dir(dir_list: list[str]):
    for ele in dir_list:
        path: Path = Path(ele)

        if not path.exists():
            path.mkdir()

def rm_temp_dir(dir_list: list[str]):
   for ele in dir_list:
       path: Path = Path(ele)

       if path.exists():
        unlink_children(path)
        path.rmdir()

def mk_files(path: Path, files: dict[str, str | dict[str, str]]):
   for key, val in files.items():
    new_path: Path = path / key

    if not new_path.exists():
        if val == "":
            new_path.touch()
        elif isinstance(val, dict):
            new_path.mkdir()
            mk_files(new_path, val)

def e_test_zip():
   files: dict[str, str | dict[str, str]] = {
      "folder1": {"file1": ""},
      "file2": "",
      "file3": "",
      "file4": ""
   }
   dir_list: list[str] = [zip_dir, pkg_dir]

   mk_temp_dir(dir_list)
   mk_files(Path(pkg_dir), files)
	
   zip_path_obj: Path = Path(zip_dir) / Vars.ZIP_FILE_NAME.value
   # needs to be relative for ziplist
   cmd: list[str] = f"zip -r {Vars.DIST_PATH.value}-temp/{Vars.ZIP_FILE_NAME.value} ./pkg-files-temp".split()
   if not zip_path_obj.exists():
      subprocess.run(cmd)

   zip_file: zipfile.ZipFile = zipfile.ZipFile(zip_path_obj)   

   pkg_list: list[str] = get_dir_list(Path(pkg_dir), replace_home=True)
   zip_list: list[str] = [] 

   for item in zip_file.filelist[1:]:
      if not item.is_dir():
         zip_list.append(item.filename)
   
   pkg_list.sort(); zip_list.sort()

   print(pkg_list, zip_list)

   rm_temp_dir(dir_list)

   assert pkg_list == zip_list