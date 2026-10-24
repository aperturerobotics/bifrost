from __future__ import annotations

from collections.abc import AsyncIterable, AsyncIterator
from typing import Protocol

from bldr.frontend import (
    frontend_pb2 as _github_com_s4wave_spacewave_bldr_frontend_frontend_pb2,
)
from bldr.web.fetch import (
    fetch_pb2 as _github_com_s4wave_spacewave_bldr_web_fetch_fetch_pb2,
)
from starpc.call import Call, CallProtocolError
from starpc.client import Client
from starpc.server import ServiceRegistry
from starpc.service import MethodDescriptor, ServiceDescriptor, bidirectional_bytes

FRONTEND_SERVICE = ServiceDescriptor(
    "bldr.frontend.Frontend",
    (
        MethodDescriptor(
            "Watch",
            _github_com_s4wave_spacewave_bldr_frontend_frontend_pb2.WatchRequest,
            _github_com_s4wave_spacewave_bldr_frontend_frontend_pb2.Event,
            False,
            True,
        ),
        MethodDescriptor(
            "Send",
            _github_com_s4wave_spacewave_bldr_frontend_frontend_pb2.SendRequest,
            _github_com_s4wave_spacewave_bldr_frontend_frontend_pb2.SendResponse,
            False,
            False,
        ),
        MethodDescriptor(
            "Fetch",
            _github_com_s4wave_spacewave_bldr_web_fetch_fetch_pb2.FetchRequest,
            _github_com_s4wave_spacewave_bldr_web_fetch_fetch_pb2.FetchResponse,
            True,
            True,
        ),
    ),
)


class FrontendClient:
    def __init__(self, client: Client, service: str | None = None) -> None:
        self._client = client
        self._service = service or "bldr.frontend.Frontend"

    async def watch(
        self,
        request: _github_com_s4wave_spacewave_bldr_frontend_frontend_pb2.WatchRequest,
    ) -> AsyncIterator[_github_com_s4wave_spacewave_bldr_frontend_frontend_pb2.Event]:
        call = await self._client.open_call(
            self._service, "Watch", request.SerializeToString(deterministic=True)
        )
        try:
            while True:
                data = await call.receive()
                if data is None:
                    return
                response = (
                    _github_com_s4wave_spacewave_bldr_frontend_frontend_pb2.Event()
                )
                response.ParseFromString(data)
                yield response
        finally:
            await call.aclose()

    async def send(
        self,
        request: _github_com_s4wave_spacewave_bldr_frontend_frontend_pb2.SendRequest,
    ) -> _github_com_s4wave_spacewave_bldr_frontend_frontend_pb2.SendResponse:
        call = await self._client.open_call(
            self._service, "Send", request.SerializeToString(deterministic=True)
        )
        try:
            data = await call.receive()
            if data is None:
                raise CallProtocolError("missing unary response")
            response = (
                _github_com_s4wave_spacewave_bldr_frontend_frontend_pb2.SendResponse()
            )
            response.ParseFromString(data)
            if await call.receive() is not None:
                raise CallProtocolError("extra unary response")
            return response
        finally:
            await call.aclose()

    async def fetch(
        self,
        requests: AsyncIterable[
            _github_com_s4wave_spacewave_bldr_web_fetch_fetch_pb2.FetchRequest
        ],
    ) -> AsyncIterator[
        _github_com_s4wave_spacewave_bldr_web_fetch_fetch_pb2.FetchResponse
    ]:
        call = await self._client.open_call(self._service, "Fetch")

        async def encoded() -> AsyncIterator[bytes]:
            async for request in requests:
                yield request.SerializeToString(deterministic=True)

        async for data in bidirectional_bytes(call, encoded()):
            response = (
                _github_com_s4wave_spacewave_bldr_web_fetch_fetch_pb2.FetchResponse()
            )
            response.ParseFromString(data)
            yield response


class FrontendServer(Protocol):
    def watch(
        self,
        request: _github_com_s4wave_spacewave_bldr_frontend_frontend_pb2.WatchRequest,
    ) -> AsyncIterator[
        _github_com_s4wave_spacewave_bldr_frontend_frontend_pb2.Event
    ]: ...
    async def send(
        self,
        request: _github_com_s4wave_spacewave_bldr_frontend_frontend_pb2.SendRequest,
    ) -> _github_com_s4wave_spacewave_bldr_frontend_frontend_pb2.SendResponse: ...
    def fetch(
        self,
        requests: AsyncIterator[
            _github_com_s4wave_spacewave_bldr_web_fetch_fetch_pb2.FetchRequest
        ],
    ) -> AsyncIterator[
        _github_com_s4wave_spacewave_bldr_web_fetch_fetch_pb2.FetchResponse
    ]: ...


def register_frontend(
    registry: ServiceRegistry,
    implementation: FrontendServer,
    service: str = "bldr.frontend.Frontend",
) -> None:
    async def watch_handler(call: Call) -> None:
        first = await call.receive()
        if first is None:
            raise CallProtocolError("missing initial request")
        request = _github_com_s4wave_spacewave_bldr_frontend_frontend_pb2.WatchRequest()
        request.ParseFromString(first)
        async for response in implementation.watch(request):
            await call.send(response.SerializeToString(deterministic=True))

    registry.register(service, "Watch", watch_handler)

    async def send_handler(call: Call) -> None:
        first = await call.receive()
        if first is None:
            raise CallProtocolError("missing initial request")
        request = _github_com_s4wave_spacewave_bldr_frontend_frontend_pb2.SendRequest()
        request.ParseFromString(first)
        response = await implementation.send(request)
        await call.send(response.SerializeToString(deterministic=True))

    registry.register(service, "Send", send_handler)

    async def fetch_handler(call: Call) -> None:
        async def requests() -> AsyncIterator[
            _github_com_s4wave_spacewave_bldr_web_fetch_fetch_pb2.FetchRequest
        ]:
            while True:
                data = await call.receive()
                if data is None:
                    return
                request = (
                    _github_com_s4wave_spacewave_bldr_web_fetch_fetch_pb2.FetchRequest()
                )
                request.ParseFromString(data)
                yield request

        async for response in implementation.fetch(requests()):
            await call.send(response.SerializeToString(deterministic=True))

    registry.register(service, "Fetch", fetch_handler)
