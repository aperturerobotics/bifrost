# HTTP Forwarding Example

Install Bifrost:

```
go install -v github.com/aperturerobotics/bifrost/cmd/bifrost
```

Start the destination service that we will proxy connections to:

```
python3 -m http.server 8080
```

Start the first peer, which forwards incoming streams to localhost:8080:

```
bifrost daemon --node-priv ../priv/node-1.pem -c node-1.yaml
```

Start the second peer, which listens on :8084 and forwards incoming traffic to
the other peer via. Bifrost:

```
bifrost daemon --node-priv ../priv/node-2.pem -c node-2.yaml
```

Access the forwarded HTTP service via the proxy:

```
curl localhost:8084
```

Or browse to http://localhost:8084 in a web browser.

When opening a connection to the second node at port 8084, Bifrost will open a
Quic-based Link with the other peer on-demand and proxy the traffic to the
destination service over a stream. When the proxy becomes idle, Bifrost will
close the Link after a short inactivity period.
