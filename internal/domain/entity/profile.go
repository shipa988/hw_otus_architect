package entity

import "context"


type Profile struct {
	*User
	IsFriend bool
}

type ProfileRepository interface {
	IsSubscribed(ctx context.Context,userID1 uint64,userID2 uint64) (bool, error)
}
