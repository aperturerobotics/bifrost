from __future__ import annotations

from collections.abc import AsyncIterable, AsyncIterator
from typing import Protocol

from rpcstream import (
    rpcstream_pb2 as _github_com_aperturerobotics_starpc_rpcstream_rpcstream_pb2,
)
from bldr.web.view import (
    view_pb2 as _github_com_s4wave_spacewave_bldr_web_view_view_pb2,
)
from starpc.call import Call, CallProtocolError
from starpc.client import Client
from starpc.server import ServiceRegistry
from starpc.service import MethodDescriptor, ServiceDescriptor, bidirectional_bytes

WEBVIEWHOST_SERVICE = ServiceDescriptor("web.view.WebViewHost", ())

WEBVIEW_SERVICE = ServiceDescriptor(
    "web.view.WebView",
    (
        MethodDescriptor(
            "SetRenderMode",
            _github_com_s4wave_spacewave_bldr_web_view_view_pb2.SetRenderModeRequest,
            _github_com_s4wave_spacewave_bldr_web_view_view_pb2.SetRenderModeResponse,
            False,
            False,
        ),
        MethodDescriptor(
            "SetHtmlLinks",
            _github_com_s4wave_spacewave_bldr_web_view_view_pb2.SetHtmlLinksRequest,
            _github_com_s4wave_spacewave_bldr_web_view_view_pb2.SetHtmlLinksResponse,
            False,
            False,
        ),
        MethodDescriptor(
            "ResetWebView",
            _github_com_s4wave_spacewave_bldr_web_view_view_pb2.ResetWebViewRequest,
            _github_com_s4wave_spacewave_bldr_web_view_view_pb2.ResetWebViewResponse,
            False,
            False,
        ),
        MethodDescriptor(
            "RemoveWebView",
            _github_com_s4wave_spacewave_bldr_web_view_view_pb2.RemoveWebViewRequest,
            _github_com_s4wave_spacewave_bldr_web_view_view_pb2.RemoveWebViewResponse,
            False,
            False,
        ),
    ),
)

ACCESSWEBVIEWS_SERVICE = ServiceDescriptor(
    "web.view.AccessWebViews",
    (
        MethodDescriptor(
            "WebViewRpc",
            _github_com_aperturerobotics_starpc_rpcstream_rpcstream_pb2.RpcStreamPacket,
            _github_com_aperturerobotics_starpc_rpcstream_rpcstream_pb2.RpcStreamPacket,
            True,
            True,
        ),
    ),
)


class WebViewHostClient:
    def __init__(self, client: Client, service: str | None = None) -> None:
        self._client = client
        self._service = service or "web.view.WebViewHost"


class WebViewClient:
    def __init__(self, client: Client, service: str | None = None) -> None:
        self._client = client
        self._service = service or "web.view.WebView"

    async def set_render_mode(
        self,
        request: _github_com_s4wave_spacewave_bldr_web_view_view_pb2.SetRenderModeRequest,
    ) -> _github_com_s4wave_spacewave_bldr_web_view_view_pb2.SetRenderModeResponse:
        call = await self._client.open_call(
            self._service,
            "SetRenderMode",
            request.SerializeToString(deterministic=True),
        )
        try:
            data = await call.receive()
            if data is None:
                raise CallProtocolError("missing unary response")
            response = _github_com_s4wave_spacewave_bldr_web_view_view_pb2.SetRenderModeResponse()
            response.ParseFromString(data)
            if await call.receive() is not None:
                raise CallProtocolError("extra unary response")
            return response
        finally:
            await call.aclose()

    async def set_html_links(
        self,
        request: _github_com_s4wave_spacewave_bldr_web_view_view_pb2.SetHtmlLinksRequest,
    ) -> _github_com_s4wave_spacewave_bldr_web_view_view_pb2.SetHtmlLinksResponse:
        call = await self._client.open_call(
            self._service, "SetHtmlLinks", request.SerializeToString(deterministic=True)
        )
        try:
            data = await call.receive()
            if data is None:
                raise CallProtocolError("missing unary response")
            response = _github_com_s4wave_spacewave_bldr_web_view_view_pb2.SetHtmlLinksResponse()
            response.ParseFromString(data)
            if await call.receive() is not None:
                raise CallProtocolError("extra unary response")
            return response
        finally:
            await call.aclose()

    async def reset_web_view(
        self,
        request: _github_com_s4wave_spacewave_bldr_web_view_view_pb2.ResetWebViewRequest,
    ) -> _github_com_s4wave_spacewave_bldr_web_view_view_pb2.ResetWebViewResponse:
        call = await self._client.open_call(
            self._service, "ResetWebView", request.SerializeToString(deterministic=True)
        )
        try:
            data = await call.receive()
            if data is None:
                raise CallProtocolError("missing unary response")
            response = _github_com_s4wave_spacewave_bldr_web_view_view_pb2.ResetWebViewResponse()
            response.ParseFromString(data)
            if await call.receive() is not None:
                raise CallProtocolError("extra unary response")
            return response
        finally:
            await call.aclose()

    async def remove_web_view(
        self,
        request: _github_com_s4wave_spacewave_bldr_web_view_view_pb2.RemoveWebViewRequest,
    ) -> _github_com_s4wave_spacewave_bldr_web_view_view_pb2.RemoveWebViewResponse:
        call = await self._client.open_call(
            self._service,
            "RemoveWebView",
            request.SerializeToString(deterministic=True),
        )
        try:
            data = await call.receive()
            if data is None:
                raise CallProtocolError("missing unary response")
            response = _github_com_s4wave_spacewave_bldr_web_view_view_pb2.RemoveWebViewResponse()
            response.ParseFromString(data)
            if await call.receive() is not None:
                raise CallProtocolError("extra unary response")
            return response
        finally:
            await call.aclose()


class AccessWebViewsClient:
    def __init__(self, client: Client, service: str | None = None) -> None:
        self._client = client
        self._service = service or "web.view.AccessWebViews"

    async def web_view_rpc(
        self,
        requests: AsyncIterable[
            _github_com_aperturerobotics_starpc_rpcstream_rpcstream_pb2.RpcStreamPacket
        ],
    ) -> AsyncIterator[
        _github_com_aperturerobotics_starpc_rpcstream_rpcstream_pb2.RpcStreamPacket
    ]:
        call = await self._client.open_call(self._service, "WebViewRpc")

        async def encoded() -> AsyncIterator[bytes]:
            async for request in requests:
                yield request.SerializeToString(deterministic=True)

        async for data in bidirectional_bytes(call, encoded()):
            response = _github_com_aperturerobotics_starpc_rpcstream_rpcstream_pb2.RpcStreamPacket()
            response.ParseFromString(data)
            yield response


class WebViewHostServer(Protocol):
    pass


def register_web_view_host(
    registry: ServiceRegistry,
    implementation: WebViewHostServer,
    service: str = "web.view.WebViewHost",
) -> None:
    pass


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
) -> None:
    async def set_render_mode_handler(call: Call) -> None:
        first = await call.receive()
        if first is None:
            raise CallProtocolError("missing initial request")
        request = (
            _github_com_s4wave_spacewave_bldr_web_view_view_pb2.SetRenderModeRequest()
        )
        request.ParseFromString(first)
        response = await implementation.set_render_mode(request)
        await call.send(response.SerializeToString(deterministic=True))

    registry.register(service, "SetRenderMode", set_render_mode_handler)

    async def set_html_links_handler(call: Call) -> None:
        first = await call.receive()
        if first is None:
            raise CallProtocolError("missing initial request")
        request = (
            _github_com_s4wave_spacewave_bldr_web_view_view_pb2.SetHtmlLinksRequest()
        )
        request.ParseFromString(first)
        response = await implementation.set_html_links(request)
        await call.send(response.SerializeToString(deterministic=True))

    registry.register(service, "SetHtmlLinks", set_html_links_handler)

    async def reset_web_view_handler(call: Call) -> None:
        first = await call.receive()
        if first is None:
            raise CallProtocolError("missing initial request")
        request = (
            _github_com_s4wave_spacewave_bldr_web_view_view_pb2.ResetWebViewRequest()
        )
        request.ParseFromString(first)
        response = await implementation.reset_web_view(request)
        await call.send(response.SerializeToString(deterministic=True))

    registry.register(service, "ResetWebView", reset_web_view_handler)

    async def remove_web_view_handler(call: Call) -> None:
        first = await call.receive()
        if first is None:
            raise CallProtocolError("missing initial request")
        request = (
            _github_com_s4wave_spacewave_bldr_web_view_view_pb2.RemoveWebViewRequest()
        )
        request.ParseFromString(first)
        response = await implementation.remove_web_view(request)
        await call.send(response.SerializeToString(deterministic=True))

    registry.register(service, "RemoveWebView", remove_web_view_handler)


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
) -> None:
    async def web_view_rpc_handler(call: Call) -> None:
        async def requests() -> AsyncIterator[
            _github_com_aperturerobotics_starpc_rpcstream_rpcstream_pb2.RpcStreamPacket
        ]:
            while True:
                data = await call.receive()
                if data is None:
                    return
                request = _github_com_aperturerobotics_starpc_rpcstream_rpcstream_pb2.RpcStreamPacket()
                request.ParseFromString(data)
                yield request

        async for response in implementation.web_view_rpc(requests()):
            await call.send(response.SerializeToString(deterministic=True))

    registry.register(service, "WebViewRpc", web_view_rpc_handler)
