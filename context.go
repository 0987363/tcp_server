package tcp_server

import (
	"math"
)

type Context struct {
	Errors []error

	handlers HandlersChain
	index    int

	engine *Engine

	Keys map[string]interface{}

	conn Connection

	cache []byte
	size  int
	cnt   int

	onConnectionOpen   func(c *Context)
	onConnectionClosed func(c *Context)
	onNewMessage       func(c *Context)

	isOpened bool
}

const abortIndex int = math.MaxInt8 / 2

func (c *Context) MsgCount() int {
	return c.cnt
}

func (c *Context) RemoteAddr() string {
	return c.conn.RemoteAddr()
}

func (c *Context) IsAborted() bool {
	return len(c.Errors) != 0
}

func (c *Context) AbortWithError(err error) {
	c.Errors = append(c.Errors, err)
}

func (c *Context) Data(data []byte) (int, error) {
	return c.conn.Send(data)
}
func (c *Context) DataWithError(data []byte, err error) {
	c.Data(data)
	c.AbortWithError(err)
}

func (c *Context) Set(key string, value interface{}) {
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[key] = value
}

func (c *Context) Get(key string) (value interface{}, exists bool) {
	value, exists = c.Keys[key]
	return
}

func (c *Context) Next() {
	c.index++
	for s := len(c.handlers); c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Context) run() {
	c.conn.Run(c)
}

func (c *Context) Recv() (error) {
	n, err := c.conn.Recv(c.cache[c.size:])
	if err != nil {
		return err
	}
	c.size += n
	c.cnt++
	return nil
}

func (c *Context) Trim(length int) {
	if length >= c.size {
		c.size = 0
		return
	}

	copy(c.cache, c.cache[length:c.size])
	c.size -= length
}

func (c *Context) ReadData() []byte {
	return c.cache[:c.size]
}

func (c *Context) Reset() {
	c.Errors = c.Errors[:0]
	c.size = 0
	c.cnt = 0
}

/*
func (c *Context) Recv() ([]byte, error) {
	c.tcpConn.SetReadDeadline(time.Now().Add(c.engine.timeout))
	n, err := c.tcpConn.Read(c.cache[c.size:])
	if err != nil {
		return nil, err
	}
	c.size += n
	return c.cache[:c.size], nil
}

func (c *Context) Send(b []byte) error {
	c.tcpConn.SetWriteDeadline(time.Now().Add(c.engine.timeout))
	_, err := c.tcpConn.Write(b)
	return err
}

func (c *Context) TcpConn() net.Conn {
	return c.tcpConn
}

func (c *Context) TcpClose() error {
	return c.tcpCconn.Close()
}
*/
