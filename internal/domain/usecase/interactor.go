package usecase

import (
	"context"
	"encoding/base64"
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/shipa988/hw_otus_architect/internal/domain/entity"
	"github.com/twinj/uuid"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"time"
)

const (
	errHash    = "can't hash pass"
	errCompare = "can't compare pass and hash"
)

const (
	errLogin = "can't login user '%v'"
)

var _ NetworkCore = (*Interactor)(nil)

type Interactor struct {
	userRepo entity.UserRepository
}

func NewInteractor(userRepo entity.UserRepository) *Interactor {
	return &Interactor{userRepo: userRepo}
}

func (i Interactor) Login(user, pass string) (string, error) {
	ctx, _ := context.WithTimeout(context.TODO(), time.Second*5)

	ok, err := i.userRepo.Validate(ctx, user, pass)
	if err != nil {
		return "", errors.Wrap(err, errLogin)
	}
	if !ok{
		return "", errors.New(errLogin)
	}

	u, err := i.userRepo.GetUser(ctx, user)
	if err != nil {
		return "", errors.Wrap(err, errLogin)
	}

	ok, err = comparePasswords(u.PassHash, pass)
	if err != nil {
		return "", errors.Wrap(err, errLogin)
	}
	if !ok{
		return "", errors.New(errLogin)
	}

	ts, err := createToken(u.Id, u.PassHash)
	if err != nil {
		return "", errors.Wrap(err, errLogin)
	}

	at := time.Unix(ts.AtExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(ts.RtExpires, 0)
	now := time.Now()

	errAccess := i.userRepo.SignIn(ctx, ts.AccessUuid, u.Id)
	if errAccess != nil {
		return "", errors.Wrap(err, errLogin)
	}
	errRefresh := i.userRepo.SignIn(ctx, ts.RefreshUuid, u.Id)
	if errRefresh != nil {
		return "", errors.Wrap(err, errLogin)
	}
	return nil

	err = createAuth(uid, ts)
	if err != nil {
		return "", errors.Wrap(err, errLogin)
	}
	jwtbase := base64.StdEncoding.EncodeToString([]byte("access_token:" + ts.AccessToken + ",refresh_token:" + ts.RefreshToken))
	return jwtbase, nil
}

func (i Interactor) Logout() error {
	panic("implement me")
}

func (i Interactor) SignUp(user, pass string) error {
	hash, err := getHash(user, pass) //get hash
	if err != nil {
		return "", errors.Wrap(err, errLogin)
	}
	panic("implement me")
}

func (i Interactor) SendProfile(name, surName string, age int, gen entity.Gender, interest string, city string) error {
	panic("implement me")
}

func (i Interactor) GetFriends(age, gen entity.Gender, limit int, lastID uint) ([]entity.User, error) {
	panic("implement me")
}

func (i Interactor) Subscribe(id uint) error {
	panic("implement me")
}

func (i Interactor) UnSubscribe(id uint) error {
	panic("implement me")
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

func createToken(userid uint64, hash string) (*TokenDetails, error) {
	td := &TokenDetails{}
	td.AtExpires = time.Now().Add(time.Minute * 15).Unix()
	td.AccessUuid = uuid.NewV4().String()

	td.RtExpires = time.Now().Add(time.Hour * 24 * 7).Unix()
	td.RefreshUuid = uuid.NewV4().String()

	var err error
	//Creating Access Token
	//os.Setenv("ACCESS_SECRET", "jdnfksdmfksd") //this should be in an env file
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["user_id"] = userid
	atClaims["exp"] = td.AtExpires
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(hash))
	if err != nil {
		return nil, err
	}
	//Creating Refresh Token
	//os.Setenv("REFRESH_SECRET", "mcmvmkmsdnfsdmfdsjf") //this should be in an env file
	rtClaims := jwt.MapClaims{}
	rtClaims["refresh_uuid"] = td.RefreshUuid
	rtClaims["user_id"] = userid
	rtClaims["exp"] = td.RtExpires
	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(hash))
	if err != nil {
		return nil, err
	}
	return td, nil
}

func createAuth(userid uint64, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	rt := time.Unix(td.RtExpires, 0)
	now := time.Now()

	errAccess := client.Set(td.AccessUuid, strconv.Itoa(int(userid)), at.Sub(now)).Err()
	if errAccess != nil {
		return errAccess
	}
	errRefresh := client.Set(td.RefreshUuid, strconv.Itoa(int(userid)), rt.Sub(now)).Err()
	if errRefresh != nil {
		return errRefresh
	}
	return nil
}

func getHash(user, pass string) (hash string, err error) {
	if user != "" && pass != "" {
		return hashAndSalt(pass)
	}
	return "", errors.New("wrong user or pass")
}

func hashAndSalt(pwd string) (string, error) {
	// Use GenerateFromPassword to hash & salt pwd.
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.MinCost)
	if err != nil {
		return "", errors.Wrap(err, errHash)
	}
	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash), nil
}

func comparePasswords(hashedPwd, plainPwd string) (bool, error) {
	// Since we'll be getting the hashed password from the DB it
	// will be a string so we'll need to convert it to a byte slice
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, []byte(plainPwd))
	if err != nil {

		return false, errors.Wrap(err, errCompare)
	}
	return true, nil
}
