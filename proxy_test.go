package proxy_test

import (
	"net"
	"testing"
	"time"

	"github.com/bbokorney/proxy"
)

var proxyAddr = "127.0.0.1:54321"
var remoteAddr = "127.0.0.1:54322"

func checkUnexpected(err error, msg string, t *testing.T) {
	if err != nil {
		t.Fatalf("Unexpected failure: %s: %s", msg, err)
	}
}

func TestProxiedTCPConn(t *testing.T) {
	// start the proxy server
	proxyServer, err := net.Listen("tcp", proxyAddr)
	checkUnexpected(err, "Error opening proxy server port", t)
	defer proxyServer.Close()

	// start the remote server
	remoteServer, err := net.Listen("tcp", remoteAddr)
	checkUnexpected(err, "Error opening remote server port", t)
	defer remoteServer.Close()

	// open connection to proxy
	proxyClient, err := net.Dial("tcp", proxyAddr)
	checkUnexpected(err, "Error opening proxy client connection", t)
	defer proxyClient.Close()

	// accept connection to proxyj
	clientConn, err := proxyServer.Accept()
	checkUnexpected(err, "Error accepting proxy client connection", t)
	defer clientConn.Close()

	// proxy the connection
	p := proxy.ProxiedTCPConn{
		LocalConn:      clientConn,
		RemoteAddr:     remoteAddr,
		ConnectTimeout: 10 * time.Second,
		IOTimeout:      10 * time.Second,
	}
	go func() {
		err := p.Proxy()
		checkUnexpected(err, "Error proxying connection", t)
	}()

	// accept the proxied connection
	remoteConn, err := remoteServer.Accept()
	checkUnexpected(err, "Error accepting proxied connection", t)
	defer remoteConn.Close()

	// write to the client
	testData := "test data"
	_, err = proxyClient.Write([]byte(testData))
	checkUnexpected(err, "Error writing to client", t)

	// check the data on the remote server
	buf := make([]byte, 128)
	n, err := remoteConn.Read(buf)
	checkUnexpected(err, "Error reading from client", t)
	data := string(buf[0:n])
	if data != testData {
		t.Fatalf("Expected %s but got %s", testData, data)
	}

	// now send a response from the server
	testResp := "test response"
	_, err = remoteConn.Write([]byte(testResp))
	checkUnexpected(err, "Error responding to client", t)

	// receive the response on the client
	n, err = proxyClient.Read(buf)
	checkUnexpected(err, "Error reading response", t)
	respData := string(buf[0:n])
	if respData != testResp {
		t.Fatalf("Expected %s but got %s", testResp, respData)
	}
}

func TestProxiedTCPConnIOTimeout(t *testing.T) {
	// start the proxy server
	proxyServer, err := net.Listen("tcp", proxyAddr)
	checkUnexpected(err, "Error opening proxy server port", t)
	defer proxyServer.Close()

	// start the remote server
	remoteServer, err := net.Listen("tcp", remoteAddr)
	checkUnexpected(err, "Error opening remote server port", t)
	defer remoteServer.Close()

	// open connection to proxy
	proxyClient, err := net.Dial("tcp", proxyAddr)
	checkUnexpected(err, "Error opening proxy client connection", t)
	defer proxyClient.Close()

	// accept connection to proxy
	clientConn, err := proxyServer.Accept()
	checkUnexpected(err, "Error accepting proxy client connection", t)
	defer clientConn.Close()

	// proxy the connection
	p := proxy.ProxiedTCPConn{
		LocalConn:      clientConn,
		RemoteAddr:     remoteAddr,
		ConnectTimeout: 10 * time.Second,
		IOTimeout:      1 * time.Second,
	}
	go func() {
		err := p.Proxy()
		if err == nil {
			t.Fatalf("Timeout error not caught")
		}
	}()

	// accept the proxied connection
	remoteConn, err := remoteServer.Accept()
	checkUnexpected(err, "Error accepting proxied connection", t)
	defer remoteConn.Close()

	// write to the client
	testData := "test data"
	_, err = proxyClient.Write([]byte(testData))
	checkUnexpected(err, "Error writing to client", t)

	// now we wait for the timeout
	time.Sleep(2 * time.Second)
}
