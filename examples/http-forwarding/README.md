# HTTP Forwarding

Install Bifrost:

```
go install -v github.com/aperturerobotics/bifrost/cmd/bifrost
```

This test includes two peers.

## Node 1

Peer ID: 12D3KooWC9dBAEoTHbEXq2aaTeFit7QVdvPcb6Yf76oGQZ6dGf8N 

```
python3 -m http.server 8080
# In a separate terminal:
bifrost daemon --node-priv ../priv/node-1.pem -c node-1.yaml
```

## Node 2

Peer ID: 12D3KooWJukwYJ46o4uYSApGHAZrjrZPLHqt3EVy1etas2KVn9RP

Port 8084 forwards through to port 8080 on node 1.

```
bifrost daemon --api-listen :5111 --node-priv ../priv/node-2.pem -c node-2.yaml
curl localhost:8084
```

