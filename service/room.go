package service

// CreateRoom creates new room
func (s *Controller) CreateRoom(roomID string, userID string, username string) error {
	err := s.repo.RegisterRoom(s.db, roomID, userID, username)
	if err != nil {
		return err
	}
	return nil
}
