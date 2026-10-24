from filter import filter_pb2 as _filter_pb2
from bldr.web.view import view_pb2 as _view_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class WebViewHandlersConfig(_message.Message):
    __slots__ = ("handlers", "web_view_id", "web_view_parent_id")
    HANDLERS_FIELD_NUMBER: _ClassVar[int]
    WEB_VIEW_ID_FIELD_NUMBER: _ClassVar[int]
    WEB_VIEW_PARENT_ID_FIELD_NUMBER: _ClassVar[int]
    handlers: _containers.RepeatedCompositeFieldContainer[WebViewHandlerConfig]
    web_view_id: _filter_pb2.StringFilter
    web_view_parent_id: _filter_pb2.StringFilter
    def __init__(self, handlers: _Optional[_Iterable[_Union[WebViewHandlerConfig, _Mapping]]] = ..., web_view_id: _Optional[_Union[_filter_pb2.StringFilter, _Mapping]] = ..., web_view_parent_id: _Optional[_Union[_filter_pb2.StringFilter, _Mapping]] = ...) -> None: ...

class WebViewHandlerConfig(_message.Message):
    __slots__ = ("web_view_id", "web_view_parent_id", "set_render_mode", "set_html_links")
    WEB_VIEW_ID_FIELD_NUMBER: _ClassVar[int]
    WEB_VIEW_PARENT_ID_FIELD_NUMBER: _ClassVar[int]
    SET_RENDER_MODE_FIELD_NUMBER: _ClassVar[int]
    SET_HTML_LINKS_FIELD_NUMBER: _ClassVar[int]
    web_view_id: _filter_pb2.StringFilter
    web_view_parent_id: _filter_pb2.StringFilter
    set_render_mode: _view_pb2.SetRenderModeRequest
    set_html_links: _view_pb2.SetHtmlLinksRequest
    def __init__(self, web_view_id: _Optional[_Union[_filter_pb2.StringFilter, _Mapping]] = ..., web_view_parent_id: _Optional[_Union[_filter_pb2.StringFilter, _Mapping]] = ..., set_render_mode: _Optional[_Union[_view_pb2.SetRenderModeRequest, _Mapping]] = ..., set_html_links: _Optional[_Union[_view_pb2.SetHtmlLinksRequest, _Mapping]] = ...) -> None: ...
