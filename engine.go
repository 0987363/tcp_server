package tcp_server

import (
	"crypto/tls"
	"log"
	"net"
	"time"
)

var engine *Engine

func init() {

}

type HandlerFunc func(*Context)
type HandlersChain []HandlerFunc

type HandlersMiddware struct {
	Handlers HandlersChain
	Index    int
}
type Engine struct {
	address   string
	config    *tls.Config
	timeout   time.Duration
	cacheSize int
	udpProc   int

	onConnectionOpen   func(c *Context)
	onConnectionClosed func(c *Context)
	onNewMessage       func(c *Context)

	handlers HandlersChain
}

func (s *Engine) Use(middleware ...HandlerFunc) {
	s.handlers = append(s.handlers, middleware...)
}

func (s *Engine) OnConnectionOpen(callback func(c *Context)) {
	s.onConnectionOpen = callback
}

func (s *Engine) OnConnectionClosed(callback func(c *Context)) {
	s.onConnectionClosed = callback
}

func (s *Engine) OnNewMessage(callback func(c *Context) ) {
	s.onNewMessage = callback
}

func (engine *Engine) NewContext(conn Connection) *Context {
	c := &Context{
		conn:   conn,
		cache:  make([]byte, engine.cacheSize),
		engine: engine,
		index:  -1,
		handlers: make(HandlersChain, len(engine.handlers) + 1),
	}
	copy(c.handlers, engine.handlers)
	c.handlers[len(c.handlers) - 1] = conn.Run

	c.onConnectionOpen = func(c *Context) {
		c.opened = true
		if engine.onConnectionOpen != nil {
			engine.onConnectionOpen(c)
		}
	}

	c.onNewMessage = func(c *Context) {
		if engine.onNewMessage != nil {
			engine.onNewMessage(c)
		}
	}

	c.onConnectionClosed = func(c *Context) {
		if engine.onConnectionClosed != nil {
			engine.onConnectionClosed(c)
		}
	}

	return c
}

func (s *Engine) Listen() {
	if s.udpProc > 0 {
		ServerAddr, err := net.ResolveUDPAddr("udp", s.address)
		if err != nil {
			log.Fatal("Resolve udp addr ", s.address, " failed:", err)
			return
		}
		conn, err := net.ListenUDP("udp", ServerAddr)
		if err != nil {
			log.Fatal("Listen udp addr ", ServerAddr, " failed:", err)
			return
		}

		for i := 0; i < s.udpProc; i++ {
			c := s.NewContext(&Udp{conn: conn})
			go c.Next()
		}
	}

	var err error
	var ln net.Listener
	if s.config == nil {
		ln, err = net.Listen("tcp", s.address)
	} else {
		ln, err = tls.Listen("tcp", s.address, s.config)
	}
	if err != nil {
		log.Fatal("Error starting TCP server.")
	}

	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Accept network failed:", err)
			continue
		}

		c := s.NewContext(&Tcp{
			conn:    conn,
			timeout: s.timeout,
		})
		go c.Next()
	}
}

func (engine *Engine) SetTimeout(t time.Duration) {
	engine.timeout = t
}

func (engine *Engine) SetCacheSize(size int) {
	engine.cacheSize = size
}

func (engine *Engine) SetUdpProc(n int) {
	engine.udpProc = n
}

func (engine *Engine) SetAddress(address string) {
	engine.address = address
}

func newEngine() *Engine {
	return &Engine{
		timeout:   time.Minute,
		cacheSize: 4096,
	}
}

// Creates new tcp server instance
func New(address string) *Engine {
	server := newEngine()
	server.address = address

	//	server.OnConnectionOpen(func(c *Context) error { return nil })
	//	server.OnNewMessage(func(c *Context, message []byte) error { return nil })
	//	server.OnConnectionClosed(func(c *Context, err error) {})

	return server
}

func NewWithTLS(address string, certFile string, keyFile string) *Engine {
	cert, _ := tls.LoadX509KeyPair(certFile, keyFile)
	config := tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	server := newEngine()
	server.address = address
	server.config = &config

	//	server.OnConnectionOpen(func(c *Context) error { return nil })
	//	server.OnNewMessage(func(c *Context, message []byte) error { return nil })
	//	server.OnConnectionClosed(func(c *Context, err error) {})

	return server
}
