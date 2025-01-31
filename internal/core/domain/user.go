package domain

type UserID uint64

type User struct {
	Login    string
	Password string
	ID       UserID
}
