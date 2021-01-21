package usecase

import "github.com/shipa988/hw_otus_architect/internal/domain/entity"

type NetworkCore interface {
	Login(user, pass string) (string,error)
	Logout() error
	SignUp(user, pass string) error
	SendProfile(name, surName string, age int, gen entity.Gender, interest string, city string) error
	GetFriends(age, gen entity.Gender, limit int, lastID uint) ([]entity.User, error)
	Subscribe(id uint) error
	UnSubscribe(id uint) error
}
