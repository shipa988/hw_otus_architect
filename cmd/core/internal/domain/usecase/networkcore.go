package usecase

import (
	"context"
	"github.com/shipa988/hw_otus_architect/internal/domain/entity"
)

type NetworkCore interface {
	Login(login, pass string) (at string, rt string, err error)
	Logout(id uint64,uuid string) error
	SignUp(login, name, pass string) (at string, rt string, err error)

	SetTokenForUser(ctx context.Context, userID uint64) (string, string, error)
	VerifyUser(token string, tokenType string) (sessionId,userId  string, err error)

	GetMyProfile(userID uint64) (entity.User, error)
	SaveMyProfile(userID uint64,name, surName string, age string, gen string, interest string, city string) error
	GetUserProfile(myUserID,otherUserId uint64) (*entity.Profile, error)

	GetFriends(userID uint64, limit int, lastID uint64) ([]entity.User, error)
	GetPeople(myuId uint64,searchName,searchSurname string, limit int, lastID uint64) ([]entity.User, error)
	Subscribe(fromId uint64,toId uint64) error
	UnSubscribe(fromId uint64,toId uint64) error

	SaveNews(myuId uint64,title,text string) (error)
	GetNews(myuId uint64) ([]entity.News, error)
	GetMyNews(myuId uint64) ([]entity.News, error)
}
