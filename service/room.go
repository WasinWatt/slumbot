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

// ListRoomIDs lists all open rooms
func (s *Controller) ListRoomIDs() ([]string, error) {
	rooms, err := s.repo.FindAllRooms(s.db)
	if err != nil {
		return []string{}, err
	}

	var roomIDs []string
	for _, r := range rooms {
		roomIDs = append(roomIDs, r.ID)
	}

	return roomIDs, nil
}

// DeleteRoom deletes a room
func (s *Controller) DeleteRoom(roomID, userID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	isAdmin, err := s.repo.IsAdmin(tx, userID)
	if err != nil {
		return err
	}

	isOwner, err := s.repo.IsOwner(tx, roomID, userID)
	if err != nil {
		return err
	}

	if !isOwner && !isAdmin {
		return ErrUnAuthorized
	}

	err = s.repo.RemoveRoomByID(tx, roomID)
	if err != nil {
		return nil
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
