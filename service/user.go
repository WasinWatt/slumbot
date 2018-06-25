package service

import (
	"database/sql"

	"github.com/WasinWatt/slumbot/user"
)

// CreateUser creates a user
func (s *Controller) CreateUser(u *user.User) error {
	err := s.repo.RegisterUser(s.db, u)
	if err != nil {
		return err
	}

	return nil
}

// Join joins user to the room
func (s *Controller) Join(u *user.User, roomID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// check if the room number exists
	exist, err := s.repo.IsRoomExists(tx, roomID)
	if err != nil {
		return err
	}

	if !exist {
		return ErrRoomNotFound
	}

	duplicate, err := s.repo.IsInRoom(tx, u.Name, roomID)
	if duplicate {
		return ErrDuplicateUserInRoom
	}
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	err = s.repo.AddToRoom(tx, u.Name, roomID)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Leave remove user from the database
func (s *Controller) Leave(u *user.User, roomID string) (int, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	duplicate, err := s.repo.IsInRoom(tx, u.Name, roomID)
	if !duplicate {
		return 0, ErrUserNotFound
	}

	if err != nil {
		return 0, err
	}

	err = s.repo.RemoveMemberByUsername(tx, u.Name, roomID)
	if err != nil {
		return 0, err
	}

	penalty, err := s.repo.AddPenalty(tx, u.ID)
	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return penalty, nil
}

// GetUser finds a user
func (s *Controller) GetUser(userID string) (*user.User, error) {
	user, err := s.repo.FindUserByID(s.db, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetAllUsernamesByRoomID finds user by room ID
func (s *Controller) GetAllUsernamesByRoomID(roomID string) ([]string, error) {
	xs, err := s.repo.FindUsernamesByRoomID(s.db, roomID)
	if err != nil {
		return []string{}, err
	}

	return xs, nil
}
