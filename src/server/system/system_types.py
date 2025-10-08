from typing import TypedDict


class LogInfo(TypedDict):
    logFileName: str
    body: str

class KeyInfo(TypedDict):
    key: str
    serialTag: str