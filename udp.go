package tcp_server

import "net"

type Udp struct {
	conn       *net.UDPConn
	remtoeAddr string
}

func (udp *Udp) RemoteAddr() string {
	return udp.remtoeAddr
}

func (udp *Udp) Run(c *Context) {
	for {
		n, remote, err := udp.recv(c.cache[c.size:])
		if err != nil {
			continue
		}
		c.size += n
		udp.remtoeAddr = remote.String()

		c.onConnectionOpen(c)
		msg, _ := c.onNewMessage(c)
		if msg != nil {
			udp.send(remote, msg)
		}
		c.onConnectionClosed(c)
	}
}

func (udp *Udp) recv(cache []byte) (int, *net.UDPAddr, error) {
	return udp.conn.ReadFromUDP(cache)
}

func (udp *Udp) send(addr *net.UDPAddr, data []byte) (int, error) {
	return udp.conn.WriteToUDP(data, addr)
}

func (udp *Udp) Close() error {
	return udp.conn.Close()
}
