## CMS-Backend
### Build command
``go build --ldflags "-s -w" main.go``

### How this API should do routing
The server implements SSL on port 443 and redirects traffic from port 80 (HTTP)
to port 443 (HTTPS). The server then listens on port 443 (HTTPS or rather WSS)
for Websocket events. Since this is a Websocket API the communication between
server and client is bidirectional.

This API handles routing over a Websocket connection by having the Websocket
listen only to the ``/socket`` endpoint. 

1. The client sends an ``Upgrade`` request to the ``/socket`` endpoint.
2. The server does all the Websocket handshake magic then replies to the
client either with a ``403 Forbidden`` if the request didn't meet the
upgrade conditions or with a ``101 Switching Protocols`` if the request
was fine.
3. Once the client recieves a ``101 Switching Protocols`` response the
Websocket channel between client and server is open and both the client
and the server will listen for incoming messages aswell as send outgoing
messages.
4. Because the server only listens on the ``/socket`` endpoint, the client
has to specify what action should be taken with every outgoing message.
If for example the client wants to request their ``id``:

Client request:
```json
{
    "action": "reqClientId",
    "payload": {}
}
```

Server response:
```json
{
    "success": "true",
    "action": "resClientId",
    "payload": {
        "clientId": "5412"
    }
}
```
