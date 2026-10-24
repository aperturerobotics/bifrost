from collections.abc import AsyncIterable, AsyncIterator
from typing import Protocol

from rpcstream import (
    rpcstream_pb2 as _github_com_aperturerobotics_starpc_rpcstream_rpcstream_pb2,
)
from bldr.web.view import (
    view_pb2 as _github_com_s4wave_spacewave_bldr_web_view_view_pb2,
)
from starpc.client import Client
from starpc.server import ServiceRegistry
from starpc.service import ServiceDescriptor

WEBVIEWHOST_SERVICE: ServiceDescriptor

class WebViewHostClient:
    def __init__(self, client: Client, service: str | None = None) -> None: ...

class WebViewHostServer(Protocol): ...

def register_web_view_host(
    registry: ServiceRegistry,
    implementation: WebViewHostServer,
    service: str = "web.view.WebViewHost",
) -> None: ...

WEBVIEW_SERVICE: ServiceDescriptor

class WebViewClient:
    def __init__(self, client: Client, service: str | None = None) -> None: ...
    async def set_render_mode(
        self,
        request: _github_com_s4wave_spacewave_bldr_web_view_view_pb2.SetRenderModeRequest,
    ) -> _github_com_s4wave_spacewave_bldr_web_view_view_pb2.SetRenderModeResponse: ...
    async def set_html_links(
        self,
        request: _github_com_s4wave_spacewave_bldr_web_view_view_pb2.SetHtmlLinksRequest,
    ) -> _github_com_s4wave_spacewave_bldr_web_view_view_pb2.SetHtmlLinksResponse: ...
    async def reset_web_view(
        self,
        request: _github_com_s4wave_spacewave_bldr_web_view_view_pb2.ResetWebViewRequest,
    ) -> _github_com_s4wave_spacewave_bldr_web_view_view_pb2.ResetWebViewResponse: ...
    async def remove_web_view(
        self,
        request: _github_com_s4wave_spacewave_bldr_web_view_view_pb2.RemoveWebViewRequest,
    ) -> _github_com_s4wave_spacewave_bldr_web_view_view_pb2.RemoveWebViewResponse: ...

class WebViewServer(Protocol):
    async def set_render_mode(
        self,
        request: _github_com_s4wave_spacewave_bldr_web_view_view_pb2.SetRenderModeRequest,
    ) -> _github_com_s4wave_spacewave_bldr_web_view_view_pb2.SetRenderModeResponse: ...
    async def set_html_links(
        self,
        request: _github_com_s4wave_spacewave_bldr_web_view_view_pb2.SetHtmlLinksRequest,
    ) -> _github_com_s4wave_spacewave_bldr_web_view_view_pb2.SetHtmlLinksResponse: ...
    async def reset_web_view(
        self,
        request: _github_com_s4wave_spacewave_bldr_web_view_view_pb2.ResetWebViewRequest,
    ) -> _github_com_s4wave_spacewave_bldr_web_view_view_pb2.ResetWebViewResponse: ...
    async def remove_web_view(
        self,
        request: _github_com_s4wave_spacewave_bldr_web_view_view_pb2.RemoveWebViewRequest,
    ) -> _github_com_s4wave_spacewave_bldr_web_view_view_pb2.RemoveWebViewResponse: ...

def register_web_view(
    registry: ServiceRegistry,
    implementation: WebViewServer,
    service: str = "web.view.WebView",
) -> None: ...

ACCESSWEBVIEWS_SERVICE: ServiceDescriptor

class AccessWebViewsClient:
    def __init__(self, client: Client, service: str | None = None) -> None: ...
    def web_view_rpc(
        self,
        requests: AsyncIterable[
            _github_com_aperturerobotics_starpc_rpcstream_rpcstream_pb2.RpcStreamPacket
        ],
    ) -> AsyncIterator[
        _github_com_aperturerobotics_starpc_rpcstream_rpcstream_pb2.RpcStreamPacket
    ]: ...

class AccessWebViewsServer(Protocol):
    def web_view_rpc(
        self,
        requests: AsyncIterator[
            _github_com_aperturerobotics_starpc_rpcstream_rpcstream_pb2.RpcStreamPacket
        ],
    ) -> AsyncIterator[
        _github_com_aperturerobotics_starpc_rpcstream_rpcstream_pb2.RpcStreamPacket
    ]: ...

def register_access_web_views(
    registry: ServiceRegistry,
    implementation: AccessWebViewsServer,
    service: str = "web.view.AccessWebViews",
) -> None: ...
