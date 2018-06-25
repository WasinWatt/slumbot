package postgres

import (
	"database/sql"
	"log"

	"github.com/WasinWatt/slumbot/sqldb"
	"github.com/WasinWatt/slumbot/user"
	"github.com/lib/pq"
)

// RegisterUser registers a user
func (r *Repository) RegisterUser(db sqldb.Queryer, u *user.User) error {
	_, err := db.Exec(`
		insert into users (
			user_id, name, penalty_num
		) values (
			$1, $2, 0
		)
		`, u.ID, u.Name,
	)
	return err
}

// IsUserExists checks if the user exists
func (r *Repository) IsUserExists(db sqldb.Queryer, userID string) (exist bool, err error) {
	err = db.QueryRow(`
		select 
			count(*) > 0 
		from 
			users
		where
			user_id = $1
		`, userID,
	).Scan(&exist)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

// FindUsernamesByRoomID finds users by room id
func (r *Repository) FindUsernamesByRoomID(db sqldb.Queryer, roomID string) (members []string, err error) {
	err = db.QueryRow(`
		select
			members
		from
			rooms
		where
			room_id = $1
		`, roomID,
	).Scan(pq.Array(&members))

	if err != nil {
		log.Println("FUCK")
		return []string{}, err
	}

	return members, nil
}

// FindUserByID finds user by user id
func (r *Repository) FindUserByID(db sqldb.Queryer, userID string) (*user.User, error) {
	var x user.User
	err := db.QueryRow(`
		select
			user_id, name, penalty_num
		from
			users
		where
			user_id = $1
		`, userID,
	).Scan(&x.ID, &x.Name, &x.PenaltyNum)

	if err != nil {
		return nil, err
	}

	return &x, nil
}

// RemoveMemberByUsername deletes member name in room
func (r *Repository) RemoveMemberByUsername(db sqldb.Queryer, username string, roomID string) error {
	_, err := db.Exec(`
		update 
			rooms
		set
			members = array_remove(members, $1)
		where
			room_id = $2
		`, username, roomID,
	)

	if err != nil {
		return err
	}

	return nil
}

// IsInRoom checks if username is already in the room
func (r *Repository) IsInRoom(db sqldb.Queryer, username string, roomID string) (duplicate bool, err error) {
	err = db.QueryRow(`
		select
			count(*) > 0
		from
			rooms
		where
			room_id = $1 and $2 = any(members)	
		`, roomID, username,
	).Scan(&duplicate)

	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return duplicate, nil
}

// AddToRoom updates room's member
func (r *Repository) AddToRoom(db sqldb.Queryer, username string, roomID string) error {
	_, err := db.Exec(`
		update 
			rooms
		set
			members = array_append(members, $1)
		where
			room_id = $2
		`, username, roomID,
	)

	if err != nil {
		return err
	}

	return nil
}

// AddPenalty increase penalty of a user by one
func (r *Repository) AddPenalty(db sqldb.Queryer, userID string) (num int, err error) {
	err = db.QueryRow(`
		update
			users
		set
			penalty_num = penalty_num + 1
		where
			user_id = $1
		
		returning penalty_num
		`, userID,
	).Scan(&num)

	return
}
