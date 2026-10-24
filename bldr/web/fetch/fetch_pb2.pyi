from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class FetchRequest(_message.Message):
    __slots__ = ("request_info", "request_data")
    REQUEST_INFO_FIELD_NUMBER: _ClassVar[int]
    REQUEST_DATA_FIELD_NUMBER: _ClassVar[int]
    request_info: FetchRequestInfo
    request_data: FetchRequestData
    def __init__(self, request_info: _Optional[_Union[FetchRequestInfo, _Mapping]] = ..., request_data: _Optional[_Union[FetchRequestData, _Mapping]] = ...) -> None: ...

class FetchRequestInfo(_message.Message):
    __slots__ = ("method", "url", "headers", "has_body", "client_id", "destination", "integrity", "mode", "redirect", "referrer", "referrer_policy")
    class HeadersEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    METHOD_FIELD_NUMBER: _ClassVar[int]
    URL_FIELD_NUMBER: _ClassVar[int]
    HEADERS_FIELD_NUMBER: _ClassVar[int]
    HAS_BODY_FIELD_NUMBER: _ClassVar[int]
    CLIENT_ID_FIELD_NUMBER: _ClassVar[int]
    DESTINATION_FIELD_NUMBER: _ClassVar[int]
    INTEGRITY_FIELD_NUMBER: _ClassVar[int]
    MODE_FIELD_NUMBER: _ClassVar[int]
    REDIRECT_FIELD_NUMBER: _ClassVar[int]
    REFERRER_FIELD_NUMBER: _ClassVar[int]
    REFERRER_POLICY_FIELD_NUMBER: _ClassVar[int]
    method: str
    url: str
    headers: _containers.ScalarMap[str, str]
    has_body: bool
    client_id: str
    destination: str
    integrity: str
    mode: str
    redirect: str
    referrer: str
    referrer_policy: str
    def __init__(self, method: _Optional[str] = ..., url: _Optional[str] = ..., headers: _Optional[_Mapping[str, str]] = ..., has_body: _Optional[bool] = ..., client_id: _Optional[str] = ..., destination: _Optional[str] = ..., integrity: _Optional[str] = ..., mode: _Optional[str] = ..., redirect: _Optional[str] = ..., referrer: _Optional[str] = ..., referrer_policy: _Optional[str] = ...) -> None: ...

class FetchRequestData(_message.Message):
    __slots__ = ("data", "done")
    DATA_FIELD_NUMBER: _ClassVar[int]
    DONE_FIELD_NUMBER: _ClassVar[int]
    data: bytes
    done: bool
    def __init__(self, data: _Optional[bytes] = ..., done: _Optional[bool] = ...) -> None: ...

class FetchResponse(_message.Message):
    __slots__ = ("response_info", "response_data")
    RESPONSE_INFO_FIELD_NUMBER: _ClassVar[int]
    RESPONSE_DATA_FIELD_NUMBER: _ClassVar[int]
    response_info: ResponseInfo
    response_data: ResponseData
    def __init__(self, response_info: _Optional[_Union[ResponseInfo, _Mapping]] = ..., response_data: _Optional[_Union[ResponseData, _Mapping]] = ...) -> None: ...

class ResponseInfo(_message.Message):
    __slots__ = ("headers", "ok", "redirected", "status", "status_text", "response_type")
    class HeadersEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    HEADERS_FIELD_NUMBER: _ClassVar[int]
    OK_FIELD_NUMBER: _ClassVar[int]
    REDIRECTED_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    STATUS_TEXT_FIELD_NUMBER: _ClassVar[int]
    RESPONSE_TYPE_FIELD_NUMBER: _ClassVar[int]
    headers: _containers.ScalarMap[str, str]
    ok: bool
    redirected: bool
    status: int
    status_text: str
    response_type: str
    def __init__(self, headers: _Optional[_Mapping[str, str]] = ..., ok: _Optional[bool] = ..., redirected: _Optional[bool] = ..., status: _Optional[int] = ..., status_text: _Optional[str] = ..., response_type: _Optional[str] = ...) -> None: ...

class ResponseData(_message.Message):
    __slots__ = ("data", "done")
    DATA_FIELD_NUMBER: _ClassVar[int]
    DONE_FIELD_NUMBER: _ClassVar[int]
    data: bytes
    done: bool
    def __init__(self, data: _Optional[bytes] = ..., done: _Optional[bool] = ...) -> None: ...
