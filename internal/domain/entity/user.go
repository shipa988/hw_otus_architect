package entity

import "context"

type Gender string

const (
	Male   Gender = "Male"
	Female Gender = "Female"
)

type User struct {
	Id       uint64
	Login    string
	PassHash string
	Name     string
	SurName  string
	Age      int
	Gen      Gender
	Interest string
	City     string
}

type UserRepository interface {
	Register(ctx context.Context, login, hash string) error
	GetUser(ctx context.Context, login string) (User,error)
	SaveUser(ctx context.Context, user User) (uint, error)
	SignIn(ctx context.Context, uuid string,id uint64) error
	Validate(ctx context.Context,login,pass string) (bool,error)
	GetUserById(ctx context.Context, id uint) (User, error)
	SaveUsersBatch(ctx context.Context, users []User) error
	FilterByNameSurName(ctx context.Context, name, surname string, limit int, lastID uint) ([]User, error)
}
