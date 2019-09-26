package tcp_server

type Connection interface {
	Run(*Context)
	RemoteAddr() string
	Send([]byte) (int, error)
	Recv([]byte) (int, error)
}
