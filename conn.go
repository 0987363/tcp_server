package tcp_server

type Connection interface {
	Run(*Context)
	RemoteAddr() string
}
