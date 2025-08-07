from typing import TypedDict

type FileTree = dict[str, str | FileTree]

class LogInfo(TypedDict):
    logFileName: str
    body: str

class Flags(TypedDict):
    zip_status: bool