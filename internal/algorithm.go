package lb

type Algorithm interface {
	GetNext(sp *ServerPool) *Server
}
