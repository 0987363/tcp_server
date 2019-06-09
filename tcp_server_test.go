package tcp_server

import (
	"net"
	"testing"
	"time"
)

func buildTestServer() *Server {
	return New("localhost:9999")
}

func Test_accepting_new_client_callback(t *testing.T) {
	server := buildTestServer()
	server.Use(func(c *Client){
		t.Error("Init logger.")
	})

//	server.OnConnectionOpen(func(c *Client) {
//		t.Error("start connection")
//	})
	server.OnNewMessage(func(c *Client, message []byte) error {
		t.Log("recv:", string(message))
		return nil
	})
	server.OnConnectionClosed(func(c *Client, err error) {
		t.Log("close, err:", err)
	})
	go server.Listen()

	// Wait for server
	// If test fails - increase this value
	time.Sleep(20 * time.Millisecond)

	conn, err := net.Dial("tcp", "localhost:9999")
	if err != nil {
		t.Fatal("Failed to connect to test server")
	}
	conn.Write([]byte("Test message\n"))
	conn.Close()

	// Wait for server
	time.Sleep(20 * time.Millisecond)
}
