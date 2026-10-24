from collections.abc import AsyncIterable, AsyncIterator
from typing import Protocol

from bldr.frontend import (
    frontend_pb2 as _github_com_s4wave_spacewave_bldr_frontend_frontend_pb2,
)
from bldr.web.fetch import (
    fetch_pb2 as _github_com_s4wave_spacewave_bldr_web_fetch_fetch_pb2,
)
from starpc.client import Client
from starpc.server import ServiceRegistry
from starpc.service import ServiceDescriptor

FRONTEND_SERVICE: ServiceDescriptor

class FrontendClient:
    def __init__(self, client: Client, service: str | None = None) -> None: ...
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
        requests: AsyncIterable[
            _github_com_s4wave_spacewave_bldr_web_fetch_fetch_pb2.FetchRequest
        ],
    ) -> AsyncIterator[
        _github_com_s4wave_spacewave_bldr_web_fetch_fetch_pb2.FetchResponse
    ]: ...

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
) -> None: ...
