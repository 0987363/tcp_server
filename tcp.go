package tcp_server

import (
	"net"
	"time"
)

type Tcp struct {
	conn       net.Conn
	timeout    time.Duration
	remtoeAddr string
}

func (tcp *Tcp) RemoteAddr() string {
	//	return tcp.conn.RemoteAddr().String()
	return tcp.remtoeAddr
}

func (tcp *Tcp) Run(c *Context) {
	defer tcp.close()
	defer c.onConnectionClosed(c)

	tcp.remtoeAddr = tcp.conn.RemoteAddr().String()

	for !c.IsAborted() {
		n, err := tcp.recv(c.cache[c.size:])
		if err != nil {
			c.AbortWithError(err)
			return
		}
		c.size += n
		if !c.isOpened {
			c.onConnectionOpen(c)
		}

		msg, err := c.onNewMessage(c)
		if err != nil {
			c.AbortWithError(err)
		}
		if msg != nil {
			if _, err := tcp.send(msg); err != nil {
				c.AbortWithError(err)
			}
		}
	}
}

func (tcp *Tcp) recv(tcpache []byte) (int, error) {
	tcp.conn.SetReadDeadline(time.Now().Add(tcp.timeout))
	return tcp.conn.Read(tcpache)
}

func (tcp *Tcp) send(data []byte) (int, error) {
	tcp.conn.SetWriteDeadline(time.Now().Add(tcp.timeout))
	return tcp.conn.Write(data)
}

func (tcp *Tcp) close() error {
	return tcp.conn.Close()
}
