from rpcstream import rpcstream_pb2 as _rpcstream_pb2
from bldr.frontend import frontend_pb2 as _frontend_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RenderMode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    RenderMode_NONE: _ClassVar[RenderMode]
    RenderMode_REACT_COMPONENT: _ClassVar[RenderMode]
    RenderMode_FUNCTION: _ClassVar[RenderMode]
    RenderMode_REACT_CHILDREN: _ClassVar[RenderMode]
RenderMode_NONE: RenderMode
RenderMode_REACT_COMPONENT: RenderMode
RenderMode_FUNCTION: RenderMode
RenderMode_REACT_CHILDREN: RenderMode

class SetRenderModeRequest(_message.Message):
    __slots__ = ("render_mode", "script_path", "props", "refresh", "frontend_binding", "web_pkg_paths")
    class WebPkgPathsEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    RENDER_MODE_FIELD_NUMBER: _ClassVar[int]
    SCRIPT_PATH_FIELD_NUMBER: _ClassVar[int]
    PROPS_FIELD_NUMBER: _ClassVar[int]
    REFRESH_FIELD_NUMBER: _ClassVar[int]
    FRONTEND_BINDING_FIELD_NUMBER: _ClassVar[int]
    WEB_PKG_PATHS_FIELD_NUMBER: _ClassVar[int]
    render_mode: RenderMode
    script_path: str
    props: bytes
    refresh: bool
    frontend_binding: _frontend_pb2.Binding
    web_pkg_paths: _containers.ScalarMap[str, str]
    def __init__(self, render_mode: _Optional[_Union[RenderMode, str]] = ..., script_path: _Optional[str] = ..., props: _Optional[bytes] = ..., refresh: _Optional[bool] = ..., frontend_binding: _Optional[_Union[_frontend_pb2.Binding, _Mapping]] = ..., web_pkg_paths: _Optional[_Mapping[str, str]] = ...) -> None: ...

class SetRenderModeResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class SetHtmlLinksRequest(_message.Message):
    __slots__ = ("clear", "remove", "set_links")
    class SetLinksEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: HtmlLink
        def __init__(self, key: _Optional[str] = ..., value: _Optional[_Union[HtmlLink, _Mapping]] = ...) -> None: ...
    CLEAR_FIELD_NUMBER: _ClassVar[int]
    REMOVE_FIELD_NUMBER: _ClassVar[int]
    SET_LINKS_FIELD_NUMBER: _ClassVar[int]
    clear: bool
    remove: _containers.RepeatedScalarFieldContainer[str]
    set_links: _containers.MessageMap[str, HtmlLink]
    def __init__(self, clear: _Optional[bool] = ..., remove: _Optional[_Iterable[str]] = ..., set_links: _Optional[_Mapping[str, HtmlLink]] = ...) -> None: ...

class HtmlLink(_message.Message):
    __slots__ = ("href", "rel")
    HREF_FIELD_NUMBER: _ClassVar[int]
    REL_FIELD_NUMBER: _ClassVar[int]
    href: str
    rel: str
    def __init__(self, href: _Optional[str] = ..., rel: _Optional[str] = ...) -> None: ...

class SetHtmlLinksResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class ResetWebViewRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class ResetWebViewResponse(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class RemoveWebViewRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class RemoveWebViewResponse(_message.Message):
    __slots__ = ("removed",)
    REMOVED_FIELD_NUMBER: _ClassVar[int]
    removed: bool
    def __init__(self, removed: _Optional[bool] = ...) -> None: ...
