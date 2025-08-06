from typing import TypedDict

type FileTree = dict[str, str | FileTree]

class Flags(TypedDict):
    zipping_status: bool