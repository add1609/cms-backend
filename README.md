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
   has to specify what action should be taken with every outgoing message.

### reqClientId
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
   "action": "resClientId",
   "success": true,
   "payload": {
      "id": "5408"
   }
}
```

### reqStartHugo
Client request:

```json
{
    "action": "reqStartHugo",
    "payload": {}
}
```

Server response:

```json
{
   "action": "resStartHugo",
   "success": true,
   "payload": {
      "previewUrl": "http://192.168.2.110:5408/preview/"
   }
}
```

### reqStopHugo
Client request:

```json
{
    "action": "reqStopHugo",
    "payload": {}
}
```

Server response:

```json
{
   "action": "resStopHugo",
   "success": true,
   "payload": {}
}
```

### reqAllFiles
Client request:

```json
{
    "action": "reqAllFiles",
    "payload": {}
}
```

Server response:

```json
{
   "action": "resAllFiles",
   "success": true,
   "payload": {
      "files": {
         "Name": "content",
         "Path": "./vielfalt/content/",
         "IsDir": true,
         "Size": 4096,
         "ModifiedTime": "2022-08-21T06:41:32.797616574+02:00",
         "Children": [
            {
               "Name": "de",
               "Path": "vielfalt/content/de",
               "IsDir": true,
               "Size": 4096,
               "ModifiedTime": "2022-08-21T06:41:32.797616574+02:00",
               "Children": [
                  {
                     "Name": "archives.md",
                     "Path": "vielfalt/content/de/archives.md",
                     "IsDir": false,
                     "Size": 43,
                     "ModifiedTime": "2022-08-21T06:41:32.797616574+02:00",
                     "Children": []
                  },
                  {
                     "Name": "categories",
                     "Path": "vielfalt/content/de/categories",
                     "IsDir": true,
                     "Size": 4096,
                     "ModifiedTime": "2022-08-21T06:41:32.797616574+02:00",
                     "Children": [
                        {
                           "Name": "_index.md",
                           "Path": "vielfalt/content/de/categories/_index.md",
                           "IsDir": false,
                           "Size": 28,
                           "ModifiedTime": "2022-08-21T06:41:32.797616574+02:00",
                           "Children": []
                        }
                     ]
                  },
                  {
                     "Name": "posts",
                     "Path": "vielfalt/content/de/posts",
                     "IsDir": true,
                     "Size": 4096,
                     "ModifiedTime": "2022-08-21T06:41:32.797616574+02:00",
                     "Children": [
                        {
                           "Name": "about",
                           "Path": "vielfalt/content/de/posts/about",
                           "IsDir": true,
                           "Size": 4096,
                           "ModifiedTime": "2022-08-21T06:41:32.797616574+02:00",
                           "Children": [
                              {
                                 "Name": "index.md",
                                 "Path": "vielfalt/content/de/posts/about/index.md",
                                 "IsDir": false,
                                 "Size": 1868,
                                 "ModifiedTime": "2022-08-21T06:41:32.797616574+02:00",
                                 "Children": []
                              }
                           ]
                        }
                     ]
                  },
                  {
                     "Name": "search.md",
                     "Path": "vielfalt/content/de/search.md",
                     "IsDir": false,
                     "Size": 41,
                     "ModifiedTime": "2022-08-21T06:41:32.797616574+02:00",
                     "Children": []
                  }
               ]
            }
         ]
      }
   }
}
```
