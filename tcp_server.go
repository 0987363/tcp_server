package tcp_server

import (
	"crypto/tls"
	"log"
	"net"
	"time"
)

// Client holds info about connection
type Client struct {
	conn net.Conn
	Keys map[string]interface{}

	cache []byte
	index int8

	Server
}

// TCP server
type Server struct {
	address string // Address to open connection: localhost:9999
	config  *tls.Config

	onConnectionOpen   func(c *Client) error
	onConnectionClosed func(c *Client, err error)
	onNewMessage       func(c *Client, message []byte) error

	Handlers HandlersChain
}

type HandlerFunc func(*Client)
type HandlersChain []HandlerFunc

const (
	ClientKey = "clientKey"
)

func (s *Server) Use(middleware ...HandlerFunc) {
	s.Handlers = append(s.Handlers, middleware...)
}

func (c *Client) Next() {
	c.index++
	for s := int8(len(c.Handlers)); c.index < s; c.index++ {
		c.Handlers[c.index](c)
	}
}

func (c *Client) listen() {
	defer c.conn.Close()

	c.Next()
	if err := c.onConnectionOpen(c); err != nil {
		return
	}

	for {
		msg, err := c.Recv()
		if err != nil {
			c.onConnectionClosed(c, err)
			return
		}

		if err := c.onNewMessage(c, msg); err != nil {
			c.onConnectionClosed(c, err)
			return
		}
	}
}

func (c *Client) Set(key string, value interface{}) {
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[key] = value
}

func (c *Client) Get(key string) (value interface{}, exists bool) {
	value, exists = c.Keys[key]
	return
}

func (c *Client) Recv() ([]byte, error) {
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	n, err := c.conn.Read(c.cache)
	if err != nil {
		return nil, err
	}
	return c.cache[:n], nil
}

func (c *Client) Send(b []byte) error {
	c.conn.SetWriteDeadline(time.Now().Add(60 * time.Second))
	_, err := c.conn.Write(b)
	return err
}

func (c *Client) Conn() net.Conn {
	return c.conn
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (s *Server) OnConnectionOpen(callback func(c *Client) error) {
	s.onConnectionOpen = callback
}

func (s *Server) OnConnectionClosed(callback func(c *Client, err error)) {
	s.onConnectionClosed = callback
}

// Called when Client receives new message
func (s *Server) OnNewMessage(callback func(c *Client, message []byte) error) {
	s.onNewMessage = callback
}

// Listen starts network server
func (s *Server) Listen(ln net.Listener) {
	if ln == nil {
		var err error
		if s.config == nil {
			ln, err = net.Listen("tcp", s.address)
		} else {
			ln, err = tls.Listen("tcp", s.address, s.config)
		}
		if err != nil {
			log.Fatal("Error starting TCP server.")
		}
	}
	defer ln.Close()

	for {
		conn, _ := ln.Accept()
		client := &Client{
			conn:   conn,
			Server: *s,
			cache:  make([]byte, 4096),
			index:  -1,
		}
		go client.listen()
	}
}

// Creates new tcp server instance
func New(address string) *Server {
	server := &Server{
		address: address,
		config:  nil,
	}

	server.OnConnectionOpen(func(c *Client) error { return nil })
	server.OnNewMessage(func(c *Client, message []byte) error { return nil })
	server.OnConnectionClosed(func(c *Client, err error) {})

	return server
}

func NewWithTLS(address string, certFile string, keyFile string) *Server {
	cert, _ := tls.LoadX509KeyPair(certFile, keyFile)
	config := tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	server := &Server{
		address: address,
		config:  &config,
	}

	server.OnConnectionOpen(func(c *Client) error { return nil })
	server.OnNewMessage(func(c *Client, message []byte) error { return nil })
	server.OnConnectionClosed(func(c *Client, err error) {})

	return server
}
