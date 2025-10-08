from typing import TypedDict

class Info(TypedDict):
    body: str

class LogInfo(Info):
    logFileName: str

class KeyInfo(Info):
    key: str
    serialTag: str