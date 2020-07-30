package tcp_server

import (
	"net"
	"testing"
	"time"
)

func Test_accepting_new_client_callback(t *testing.T) {
	server := New("localhost:9999")
	server.Use(func(c *Context) {
		t.Error("Init logger.")
	})

	//	server.OnConnectionOpen(func(c *Context) {
	//		t.Error("start connection")
	//	})
	server.OnNewMessage(func(c *Context) {
		c.Recv()
		t.Log("recv:", string(c.ReadData()))
	})
	server.OnConnectionClosed(func(c *Context) {
		t.Log("close, err:", c.Errors)
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

func Test_accepting_new_client_connection_callback(t *testing.T) {
	server := New("localhost:9998")
	server.Use(func(c *Context) {
		t.Error("Init logger.")
	})

	server.OnConnectionOpen(func(c *Context) {
		t.Log("new connection.")
	})
	server.OnConnectionClosed(func(c *Context) {
		t.Log("close, err:", c.Errors)
	})
	go server.Listen()

	// Wait for server
	// If test fails - increase this value
	time.Sleep(20 * time.Millisecond)

	conn, err := net.Dial("tcp", "localhost:9998")
	if err != nil {
		t.Fatal("Failed to connect to test server")
	}
	conn.Write([]byte("Test message\n"))
	conn.Close()

	// Wait for server
	time.Sleep(20 * time.Millisecond)
}
