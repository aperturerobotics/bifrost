set -e

go build -o chat-client ./main.go

echo "Chat client built successfully"
echo "Usage: ./chat-client --peer-id <remote-peer-id> --username <your-username>"
