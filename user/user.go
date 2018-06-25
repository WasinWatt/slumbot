package user

// User is the user struct
type User struct {
	ID         string `db:"user_id"`
	Name       string `db:"name"`
	PenaltyNum int    `db:"penalty_num"`
}
