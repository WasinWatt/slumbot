package service

// CreateRoom creates new room
func (s *Controller) CreateRoom(roomID string, userID string) error {
	exist, _ := s.repo.IsRoomExists(s.db, roomID)
	if exist {
		return ErrDuplicateRoom
	}

	err := s.repo.RegisterRoom(s.db, roomID, userID)
	if err != nil {
		return err
	}
	return nil
}
