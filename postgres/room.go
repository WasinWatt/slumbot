package postgres

import (
	"github.com/WasinWatt/slumbot/room"
	"github.com/WasinWatt/slumbot/sqldb"
	"github.com/lib/pq"
)

// IsRoomExists checks if room id exist
func (r *Repository) IsRoomExists(db sqldb.Queryer, roomID string) (exist bool, err error) {
	err = db.QueryRow(`
			select 
				count(*) > 0 
			from 
				rooms 
			where
				room_id = $1
		`, roomID,
	).Scan(&exist)

	if err != nil {
		return false, err
	}

	return exist, nil
}

// RegisterRoom registers a room
func (r *Repository) RegisterRoom(db sqldb.Queryer, roomID string, userID string) error {
	_, err := db.Exec(`
			insert into rooms (
				room_id, owner_id, members
			) values (
				$1, $2, $3
			)
		`, roomID, userID, pq.Array([]string{userID}),
	)

	return err
}

// RemoveRoomByID deletes room by id
func (r *Repository) RemoveRoomByID(db sqldb.Queryer, roomID string) error {
	_, err := db.Exec(`
			delete from rooms where
				room_id = $1 
		`, roomID,
	)

	return err
}

// FindAllRooms finds all rooms
func (r *Repository) FindAllRooms(db sqldb.Queryer) ([]*room.Room, error) {
	rows, err := db.Query(`
			select room_id, owner_id, members from rooms
		`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := make([]*room.Room, 0)
	for rows.Next() {
		var r room.Room
		err := rows.Scan(&r.ID, &r.OwnerID, pq.Array(&r.Members))
		if err != nil {
			return nil, err
		}

		rooms = append(rooms, &r)
	}
	return rooms, nil
}
