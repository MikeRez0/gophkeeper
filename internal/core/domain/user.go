package domain

type UserID uint64

// User stores information about user (login, hashed password).
type User struct {
	Login    string
	Password string
	ID       UserID
}
