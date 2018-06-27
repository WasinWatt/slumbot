package service

import (
	"database/sql"
	"errors"

	"github.com/WasinWatt/slumbot/postgres"
)

// Errors
var (
	ErrDuplicateUserInRoom = errors.New("api/join: user already in the room")
	ErrRoomNotFound        = errors.New("api: room not found")
	ErrDuplicateRoom       = errors.New("api: room already exists")
	ErrUserNotInRoom       = errors.New("api: user not in room")
)

// New creates new service controller
func New(db *sql.DB, repo *postgres.Repository) *Controller {
	return &Controller{db, repo}
}

// Controller controls business flows
type Controller struct {
	db   *sql.DB
	repo *postgres.Repository
}
