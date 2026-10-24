from __future__ import annotations

from collections.abc import AsyncIterable, AsyncIterator
from typing import Protocol

from bldr.web.fetch import (
    fetch_pb2 as _github_com_s4wave_spacewave_bldr_web_fetch_fetch_pb2,
)
from starpc.call import Call
from starpc.client import Client
from starpc.server import ServiceRegistry
from starpc.service import MethodDescriptor, ServiceDescriptor, bidirectional_bytes

FETCHSERVICE_SERVICE = ServiceDescriptor(
    "web.fetch.FetchService",
    (
        MethodDescriptor(
            "Fetch",
            _github_com_s4wave_spacewave_bldr_web_fetch_fetch_pb2.FetchRequest,
            _github_com_s4wave_spacewave_bldr_web_fetch_fetch_pb2.FetchResponse,
            True,
            True,
        ),
    ),
)


class FetchServiceClient:
    def __init__(self, client: Client, service: str | None = None) -> None:
        self._client = client
        self._service = service or "web.fetch.FetchService"

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


class FetchServiceServer(Protocol):
    def fetch(
        self,
        requests: AsyncIterator[
            _github_com_s4wave_spacewave_bldr_web_fetch_fetch_pb2.FetchRequest
        ],
    ) -> AsyncIterator[
        _github_com_s4wave_spacewave_bldr_web_fetch_fetch_pb2.FetchResponse
    ]: ...


def register_fetch_service(
    registry: ServiceRegistry,
    implementation: FetchServiceServer,
    service: str = "web.fetch.FetchService",
) -> None:
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
