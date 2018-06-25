package room

// Room is the room struct
type Room struct {
	OwnerID string   `json:"owner_id"`
	ID      string   `json:"room_id"`
	Members []string `json:"members"`
}
