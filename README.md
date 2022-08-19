## CMS-Backend

### Build command

``go build --ldflags "-s -w" main.go``

### How this API handles requests

The server implements SSL on port 443 and redirects traffic from port 80 (HTTP)
to port 443 (HTTPS). The server then listens on port 443 (HTTPS or rather WSS)
for Websocket events. Since this is a Websocket API the communication between
server and client is bidirectional.

This API handles routing over a Websocket connection by having the server listen
only to the ``/socket`` endpoint. The client communicates with the server in the
following manner:

1. The client sends an ``Upgrade`` request to the ``/socket`` endpoint.
2. The server checks whether or not the request's `Origin` header matches
   `WS_CHECK_ORIGIN_HOST`, which can be set in the `.env.local` file. If they
   match, the server replies to the client with a ``101 Switching Protocols``.
   If they don't match, the server replies with a ``403 Forbidden``.
3. Once the client receives a ``101 Switching Protocols`` response the Websocket
   channel between client and server is open and both the client and the server will
   listen for incoming messages aswell as send outgoing messages.
4. Because the server only listens on the ``/socket`` endpoint, the client
   has to specify what action should be taken with every outgoing message. If
   for example the client wants to request their ``id``:

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
