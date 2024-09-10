package lb

func (s *ServerPool) SetServerHealth(serverURL string, status bool) {
	s.poolLookup[serverURL].mux.Lock()
	defer s.poolLookup[serverURL].mux.Unlock()
	s.poolLookup[serverURL].Healthy = status
}
