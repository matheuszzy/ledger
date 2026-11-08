package service

func (s *Service) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.repository.Reset()
}
