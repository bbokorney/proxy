package proxy

import (
	"io"
	"net"
	"time"
)

// ProxiedTCPConn is a proxied TCP connection
type ProxiedTCPConn struct {
	// LocalConn is the TCP connection to proxy.
	LocalConn net.Conn

	// RemoteAddr is the address of the remote server
	// to which this connection will be proxied.
	RemoteAddr string

	// ConnectTimeout specifies how long the proxy will wait while
	// connecting to the remote server before failing
	ConnectTimeout time.Duration

	// IOTimeout specifies how long the proxy will wait
	// in between each read and write before failing
	IOTimeout time.Duration
}

// Proxy starts proxying the connection and
// blocks until the connections are both closed.
func (p *ProxiedTCPConn) Proxy() error {
	// open a connection to the remote
	remote, err := net.DialTimeout("tcp", p.RemoteAddr, p.ConnectTimeout)
	if err != nil {
		return err
	}

	errChan := make(chan error)

	// copy from local -> remote
	go copy(p.LocalConn, remote, errChan, p.IOTimeout)
	// copy from remote -> local
	go copy(remote, p.LocalConn, errChan, p.IOTimeout)

	// wait for one of the connections to end
	err = <-errChan
	if err == io.EOF {
		// one of the connections closed
		// clean things up
		<-errChan // make sure to read out error from other goroutine
		return nil
	}
	// unexpected error
	<-errChan // make sure to read out error from other goroutine
	return err
}

func copy(source, sink net.Conn, errChan chan error, timeout time.Duration) {
	buf := make([]byte, 1024)
	for {
		source.SetReadDeadline(time.Now().Add(timeout))
		n, err := source.Read(buf)
		if err != nil {
			source.Close()
			errChan <- err
			return
		}

		sink.SetWriteDeadline(time.Now().Add(timeout))
		n, err = sink.Write(buf[0:n])
		if err != nil {
			sink.Close()
			errChan <- err
			return
		}
	}
}
