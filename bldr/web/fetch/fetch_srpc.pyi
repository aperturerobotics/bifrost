from collections.abc import AsyncIterable, AsyncIterator
from typing import Protocol

from bldr.web.fetch import (
    fetch_pb2 as _github_com_s4wave_spacewave_bldr_web_fetch_fetch_pb2,
)
from starpc.client import Client
from starpc.server import ServiceRegistry
from starpc.service import ServiceDescriptor

FETCHSERVICE_SERVICE: ServiceDescriptor

class FetchServiceClient:
    def __init__(self, client: Client, service: str | None = None) -> None: ...
    def fetch(
        self,
        requests: AsyncIterable[
            _github_com_s4wave_spacewave_bldr_web_fetch_fetch_pb2.FetchRequest
        ],
    ) -> AsyncIterator[
        _github_com_s4wave_spacewave_bldr_web_fetch_fetch_pb2.FetchResponse
    ]: ...

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
) -> None: ...
