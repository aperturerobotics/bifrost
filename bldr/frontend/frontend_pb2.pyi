from bldr.web.fetch import fetch_pb2 as _fetch_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Binding(_message.Message):
    __slots__ = ("entrypoint",)
    ENTRYPOINT_FIELD_NUMBER: _ClassVar[int]
    entrypoint: str
    def __init__(self, entrypoint: _Optional[str] = ...) -> None: ...

class Session(_message.Message):
    __slots__ = ("id", "route_prefix", "entrypoints", "vite_version")
    ID_FIELD_NUMBER: _ClassVar[int]
    ROUTE_PREFIX_FIELD_NUMBER: _ClassVar[int]
    ENTRYPOINTS_FIELD_NUMBER: _ClassVar[int]
    VITE_VERSION_FIELD_NUMBER: _ClassVar[int]
    id: str
    route_prefix: str
    entrypoints: _containers.RepeatedScalarFieldContainer[str]
    vite_version: str
    def __init__(self, id: _Optional[str] = ..., route_prefix: _Optional[str] = ..., entrypoints: _Optional[_Iterable[str]] = ..., vite_version: _Optional[str] = ...) -> None: ...

class WatchRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class Event(_message.Message):
    __slots__ = ("session", "sequence", "payload")
    SESSION_FIELD_NUMBER: _ClassVar[int]
    SEQUENCE_FIELD_NUMBER: _ClassVar[int]
    PAYLOAD_FIELD_NUMBER: _ClassVar[int]
    session: Session
    sequence: int
    payload: str
    def __init__(self, session: _Optional[_Union[Session, _Mapping]] = ..., sequence: _Optional[int] = ..., payload: _Optional[str] = ...) -> None: ...

class SendRequest(_message.Message):
    __slots__ = ("session_id", "payload")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    PAYLOAD_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    payload: str
    def __init__(self, session_id: _Optional[str] = ..., payload: _Optional[str] = ...) -> None: ...

class SendResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...
