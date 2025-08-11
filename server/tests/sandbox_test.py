from pathlib import Path
import zipfile

zip_path: Path = Path("deploy-zip/deploy.zip")
pkg_path: Path = Path("pkg-files")

zip_file: zipfile.ZipFile = zipfile.ZipFile(zip_path, "a")

if len(zip_file.filelist) == 0:
    for path, _, file_list in pkg_path.walk():
        for file in file_list:
            file_path: Path = Path(f"{path}/{file}")

            if file_path.exists():
                zip_file.write(file_path)
else:
    for file in zip_file.filelist:
        print(file.filename)