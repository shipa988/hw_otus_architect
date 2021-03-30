package entity

import "context"

type Gender string

const (
	Male   Gender = "Male"
	Female Gender = "Female"
	Other  Gender = "Other"
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
	GetUserById(ctx context.Context, id uint64) (User, error)
	GetFriendsById(ctx context.Context, id uint64, limit int, lastID uint64) ([]User, error)
	GetSubscribersIdById(ctx context.Context, id uint64) ([]uint64, error)
	GetUserAuth(ctx context.Context, login string) (uint64,string,error)
	SaveUser(ctx context.Context, user User) error
	Validate(ctx context.Context, login, pass string) (bool, error)
	Subscribe(ctx context.Context,fromId uint64, toId uint64) error
	UnSubscribe(ctx context.Context,fromId uint64, toId uint64) error
	FilterByNameSurName(ctx context.Context,myuId uint64, name, surname string, limit int, lastID uint64) ([]User, error)
}

type UserAuthRepository interface {
	Register(ctx context.Context, login,name, hash string) (uint64, error)
	SignIn(ctx context.Context, uuid string, id uint64) error
	IsSignIn(ctx context.Context, uuid string) (id uint64, ok bool, err error)
	LogOff(ctx context.Context, id uint64,uuid string) (err error)
}
