# proxy
A Golang library for network proxies.

Currently it supports proxying TCP connections.

## Usage

```go
// make the
p := proxy.ProxiedTCPConn{
    // the TCP connection to proxy
	LocalConn:      clientConn,

    // The address of the remote server
   RemoteAddr:      remoteAddr,

    // The timeout to connect to the remote server
	ConnectTimeout: 10 * time.Second,

    // The timeout for I/O operations
	IOTimeout:      5 * time.Second,
}

// this will block while
// the connections remain open
err := p.Proxy()
```

You may want to start the proxy in a new goroutine.
```go
go func() {
    err := p.Proxy()
}()
```
