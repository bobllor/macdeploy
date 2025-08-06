from typing import TypedDict

type FileTree = dict[str, str | FileTree]

class Flags(TypedDict):
    zip_status: bool