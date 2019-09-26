package tcp_server

import (
	"net"
)

type Udp struct {
	conn *net.UDPConn
	//	remtoeAddr string
	remote *net.UDPAddr
}

func (udp *Udp) RemoteAddr() string {
	return udp.remote.String()
}

func (udp *Udp) Run(c *Context) {
	for {
		if err := c.Recv(); err != nil {
			c.AbortWithError(err)
			return
		}

		c.onConnectionOpen(c)
		c.onNewMessage(c)
		c.onConnectionClosed(c)

		c.Reset()
	}
}

func (udp *Udp) Recv(cache []byte) (int, error) {
	n, remote, err := udp.conn.ReadFromUDP(cache)
	udp.remote = remote
	return n, err
}

func (udp *Udp) Send(data []byte) (int, error) {
	return udp.conn.WriteToUDP(data, udp.remote)
}

func (udp *Udp) Close() error {
	return udp.conn.Close()
}
