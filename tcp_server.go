package tcp_server

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"runtime/debug"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
)

// Client holds info about connection
type Client struct {
	conn net.Conn
	ctx  context.Context
	lock sync.RWMutex

	cache []byte
	size  int
	index int

	requestID  string
	remoteAddr string

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
	ClientKey = "Client"
	LoggerKey = "Logger"
)

func (s *Server) Use(middleware ...HandlerFunc) {
	s.Handlers = append(s.Handlers, middleware...)
}

func (c *Client) Next() {
	c.index++
	for s := len(c.Handlers); c.index < s; c.index++ {
		c.Handlers[c.index](c)
	}
}

func (c *Client) listen() (err error) {
	defer func() {
		c.conn.Close()
		c.onConnectionClosed(c, err)
	}()

	c.Next()

	for {
		msg, err := c.Recv()
		if err != nil {
			return err
		}

		if err := c.onNewMessage(c, msg); err != nil {
			return err
		}
	}
}

func (c *Client) Trim(length int) {
	if length > c.size {
		c.size = 0
		return
	}

	copy(c.cache, c.cache[length:c.size])
	c.size -= length
}

func (c *Client) Recv() ([]byte, error) {
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	n, err := c.conn.Read(c.cache[c.size:])
	if err != nil {
		return nil, err
	}
	c.size += n
	return c.cache[:c.size], nil
}

func (c *Client) Send(b []byte) error {
	c.conn.SetWriteDeadline(time.Now().Add(60 * time.Second))
	_, err := c.conn.Write(b)
	return err
}

func (c *Client) RemoteAddr() string {
	return c.remoteAddr
}

func (c *Client) RequestID() string {
	return c.requestID
}

func (c *Client) Set(key string, value interface{}) {
	c.lock.Lock()
	c.ctx = context.WithValue(c.ctx, key, value)
	c.lock.Unlock()
}

func (c *Client) Get(key string) interface{} {
	c.lock.RLock()
	v := c.ctx.Value(key)
	c.lock.RUnlock()
	return v
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
	s.onNewMessage = func(c *Client, message []byte) error {
		logger, ok := c.Get(LoggerKey).(*logrus.Logger)
		if !ok {
			logger = logrus.New()
		}

		defer func() {
			if result := recover(); result != nil {
				logger.Errorf("recv stack: %s\n%s\n", result, string(debug.Stack()))
			}
		}()

		return callback(c, message)
	}
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
			conn:       conn,
			Server:     *s,
			cache:      make([]byte, 4096),
			index:      -1,
			ctx:        context.Background(),
			requestID:  uuid.NewV4().String(),
			remoteAddr: conn.RemoteAddr().String(),
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
