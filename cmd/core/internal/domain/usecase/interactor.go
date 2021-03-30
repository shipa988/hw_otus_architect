package usecase

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/shipa988/hw_otus_architect/cmd/core/internal/data/controller/grpcclient"
	"github.com/shipa988/hw_otus_architect/internal/data/controller/log"
	"github.com/shipa988/hw_otus_architect/internal/domain/entity"
	"github.com/shipa988/hw_otus_architect/internal/domain/usecases"
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
	errLogin         = "can't login user '%v'"
	errLogOff        = "can't logoff user '%v'"
	errSignUp        = "can't signup user '%v'"
	errGetProfile    = "can't get user profile by id '%v'"
	errGetFriends    = "can't get friends for id '%v'"
	errGetPeople     = "can't get peolple for by %v and %v"
	errSubscribe     = "can't subscribe user id '%v' to user id '%v'"
	errUnSubscribe   = "can't unsubscribe user id '%v' to user id '%v'"
	errSaveProfile   = "can't set user profile for id '%v'"
	errVerifyToken   = "can't verify token '%v'"
	errGenerateToken = "can't generate token"
	errSaveNews      = "can't save news for user %v"
	errGetMyNews     = "can't get news of user %v"
)

var (
	ACCESS_SECRET  = []byte("BA988091D779C3202AF3B7217ABD2641")
	REFRESH_SECRET = []byte("32CF0917D16161DD6CB95BEAF12FA689")
)

var _ NetworkCore = (*Interactor)(nil)

type Interactor struct {
	userRepo     entity.UserRepository
	newsRepo     entity.NewsRepository
	newsQueue    usecases.NewsPulisher
	newsService  *grpcclient.GRPCClient
	profileRepo  entity.ProfileRepository
	userAuthRepo entity.UserAuthRepository
	ctxTimeoutS  time.Duration
}

func (i Interactor) SaveNews(myuId uint64, title, text string) error {
	ctx, _ := context.WithTimeout(context.TODO(), i.ctxTimeoutS)
	n := entity.News{
		AuthorId: myuId,
		Title:    title,
		Time:     time.Now(),
		Text:     text,
	}
	err:=i.newsRepo.SaveNews(ctx,myuId,title,text)
	if err != nil {
		return errors.Wrapf(err, errSaveNews, myuId)
	}
	err = i.newsQueue.SaveNews(ctx, n)
	if err != nil {
		return errors.Wrapf(err, errSaveNews, myuId)
	}
	return nil
}

func (i Interactor) GetNews(myuId uint64) ([]entity.News, error) {
	news, err := i.newsService.GetNews(myuId)
	if err != nil {
		return nil, err
	}
	return news, nil
}

func (i Interactor) GetMyNews(myuId uint64) ([]entity.News, error) {
	ctx, _ := context.WithTimeout(context.TODO(), i.ctxTimeoutS)
	news, err := i.newsRepo.GetNews(ctx, myuId, 1000)
	if err != nil {
		return nil, errors.Wrapf(err, errGetMyNews, myuId)
	}
	return news, nil
}

func NewInteractor(userRepo entity.UserRepository, newsRepo entity.NewsRepository, profileRepo entity.ProfileRepository, userAuthRepo entity.UserAuthRepository, newsQueue usecases.NewsPulisher, newsService *grpcclient.GRPCClient, ctxTimeoutS int) *Interactor {
	return &Interactor{userRepo: userRepo, newsRepo:newsRepo, profileRepo: profileRepo, userAuthRepo: userAuthRepo, newsQueue: newsQueue, newsService: newsService, ctxTimeoutS: time.Second * time.Duration(ctxTimeoutS)}
}

func (i Interactor) GetFriends(userID uint64, limit int, lastID uint64) ([]entity.User, error) {
	ctx, _ := context.WithTimeout(context.TODO(), i.ctxTimeoutS)
	friends, err := i.userRepo.GetFriendsById(ctx, userID, limit, lastID)
	if err != nil {
		return nil, errors.Wrapf(err, errGetFriends, userID)
	}
	return friends, nil
}

func (i Interactor) GetPeople(myuId uint64, searchName, searchSurname string, limit int, lastID uint64) ([]entity.User, error) {
	ctx, _ := context.WithTimeout(context.TODO(), i.ctxTimeoutS)
	people, err := i.userRepo.FilterByNameSurName(ctx, myuId, searchName, searchSurname, limit, lastID)
	if err != nil {
		return nil, errors.Wrapf(err, errGetPeople, searchName, searchSurname)
	}
	return people, nil
}

func (i Interactor) Subscribe(fromId uint64, toId uint64) error {
	ctx, _ := context.WithTimeout(context.TODO(), i.ctxTimeoutS)
	err := i.userRepo.Subscribe(ctx, fromId, toId)
	if err != nil {
		return errors.Wrapf(err, errSubscribe, fromId, toId)
	}
	return nil
}

func (i Interactor) UnSubscribe(fromId uint64, toId uint64) error {
	ctx, _ := context.WithTimeout(context.TODO(), i.ctxTimeoutS)
	err := i.userRepo.UnSubscribe(ctx, fromId, toId)
	if err != nil {
		return errors.Wrapf(err, errUnSubscribe, fromId, toId)
	}
	return nil
}
func (i Interactor) GetMyProfile(userID uint64) (entity.User, error) {
	ctx, _ := context.WithTimeout(context.TODO(), i.ctxTimeoutS)
	user, err := i.userRepo.GetUserById(ctx, userID)
	if err != nil {
		return user, errors.Wrapf(err, errGetProfile)
	}
	return user, nil
}

func (i Interactor) GetUserProfile(myUserID, otherUserId uint64) (*entity.Profile, error) {
	ctx, _ := context.WithTimeout(context.TODO(), i.ctxTimeoutS)
	user, err := i.userRepo.GetUserById(ctx, otherUserId)
	if err != nil {
		return nil, errors.Wrapf(err, errGetProfile, otherUserId)
	}
	isfriend, err := i.profileRepo.IsSubscribed(ctx, myUserID, otherUserId)
	profile := entity.Profile{
		User:     &user,
		IsFriend: isfriend,
	}
	return &profile, nil
}

func (i Interactor) SaveMyProfile(userID uint64, name, surName string, age string, gen string, interest string, city string) error {
	ctx, _ := context.WithTimeout(context.TODO(), i.ctxTimeoutS)
	a, err := strconv.Atoi(age)
	if err != nil {
		log.Error(errors.Wrapf(err, errSaveProfile, userID))
	}
	g := entity.Other
	switch gen {
	case "male":
		g = entity.Male
	case "female":
		g = entity.Female
	case "other":
		g = entity.Other
	default:
		log.Error(errors.Wrapf(errors.New("unknown gender"), errSaveProfile, userID))
	}
	user := entity.User{
		Id:       userID,
		Name:     name,
		SurName:  surName,
		Age:      a,
		Gen:      g,
		Interest: interest,
		City:     city,
	}
	err = i.userRepo.SaveUser(ctx, user)
	if err != nil {
		return errors.Wrapf(err, errSaveProfile, userID)
	}
	return nil
}

func (i Interactor) Logout(id uint64, uuid string) error {
	ctx, _ := context.WithTimeout(context.TODO(), i.ctxTimeoutS)
	if err := i.userAuthRepo.LogOff(ctx, id, uuid); err != nil {
		return errors.Wrap(err, errLogOff)
	}
	return nil
}

func (i Interactor) VerifyUser(token string, tokenType string) (sessionId, userId string, err error) {
	ctx, _ := context.WithTimeout(context.TODO(), i.ctxTimeoutS)
	var secret []byte
	var uuidKey = ""
	switch tokenType {
	case "at":
		secret = ACCESS_SECRET
		uuidKey = "access_uuid"
	case "rt":
		secret = REFRESH_SECRET
		uuidKey = "refresh_uuid"
	default:
		return "", "", errors.New("token type is invalid")
	}

	tk, err := verifyToken(token, secret)
	if err != nil {
		return "", "", errors.Wrapf(err, errVerifyToken, token)
	}
	claims, ok := tk.Claims.(jwt.MapClaims)
	if ok && tk.Valid {

		uuid, ok := claims[uuidKey].(string)
		if !ok {
			return "", "", errors.Wrapf(errors.New("user uuid is absent"), errVerifyToken, token)
		}
		userID, ok := claims["user_id"].(float64)
		if !ok {
			return "", "", errors.Wrapf(errors.New("user id is absent"), errVerifyToken, token)
		}

		dbUserID, isSignIn, err := i.userAuthRepo.IsSignIn(ctx, uuid)
		if err != nil {
			return "", "", errors.Wrapf(err, errVerifyToken, token)
		}
		if !isSignIn {
			return "", "", errors.Wrapf(errors.New("token is expired or not found"), errVerifyToken, token)
		}
		if dbUserID != uint64(userID) {
			return "", "", errors.Wrapf(errors.New("user id in db and in token is not equal"), errVerifyToken, token)
		}
		return uuid, strconv.FormatUint(dbUserID, 10), err
	}
	return "", "", errors.New(errVerifyToken)
}

func (i Interactor) Login(login, pass string) (at string, rt string, err error) {
	ctx, _ := context.WithTimeout(context.TODO(), i.ctxTimeoutS)

	id, phash, err := i.userRepo.GetUserAuth(ctx, login)
	if err != nil {
		return "", "", errors.Wrapf(err, errLogin, login)
	}

	ok, err := comparePasswords(phash, pass)
	if err != nil {
		return "", "", errors.Wrapf(err, errLogin, login)
	}
	if !ok {
		return "", "", errors.New(errLogin)
	}

	at, rt, err = i.SetTokenForUser(ctx, id)
	if err != nil {
		return "", "", errors.New(errLogin)
	}
	log.Info("user %v logged in", login)
	return
}

func (i Interactor) SetTokenForUser(ctx context.Context, userID uint64) (string, string, error) {
	ts, err := createToken(userID)
	if err != nil {
		return "", "", errors.New(errGenerateToken)
	}

	errAccess := i.userAuthRepo.SignIn(ctx, ts.AccessUuid, userID /*, ts.AtExpires*/)
	if errAccess != nil {
		return "", "", errors.New(errGenerateToken)
	}
	errRefresh := i.userAuthRepo.SignIn(ctx, ts.RefreshUuid, userID /*, ts.RtExpires*/)
	if errRefresh != nil {
		return "", "", errors.New(errGenerateToken)
	}
	at := base64.StdEncoding.EncodeToString([]byte(ts.AccessToken))
	rt := base64.StdEncoding.EncodeToString([]byte(ts.RefreshToken))
	return at, rt, nil
}

func (i Interactor) SignUp(login, name, pass string) (at string, rt string, err error) {
	ctx, _ := context.WithTimeout(context.TODO(), i.ctxTimeoutS)

	ok, err := i.userRepo.Validate(ctx, login, pass)
	if err != nil {
		return "", "", errors.Wrapf(err, errSignUp, login)
	}
	if !ok {
		return "", "", fmt.Errorf(errSignUp, login)
	}

	hash, err :=
		hashAndSalt(pass)
	if err != nil {
		return "", "", errors.Wrap(err, errSignUp)
	}

	id, err := i.userAuthRepo.Register(ctx, login, name, hash)
	if err != nil {
		return "", "", errors.Wrap(err, errLogin)
	}

	ts, err := createToken(id)
	if err != nil {
		return "", "", errors.Wrap(err, errLogin)
	}

	errAccess := i.userAuthRepo.SignIn(ctx, ts.AccessUuid, id /*, ts.AtExpires*/)
	if errAccess != nil {
		return "", "", errors.Wrap(err, errLogin)
	}
	errRefresh := i.userAuthRepo.SignIn(ctx, ts.RefreshUuid, id /*, ts.RtExpires*/)
	if errRefresh != nil {
		return "", "", errors.Wrap(err, errLogin)
	}
	at = base64.StdEncoding.EncodeToString([]byte(ts.AccessToken))
	rt = base64.StdEncoding.EncodeToString([]byte(ts.RefreshToken))
	log.Info("user %v signed up", login)
	return at, rt, nil
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

func createToken(userid uint64) (*TokenDetails, error) {
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
	td.AccessToken, err = at.SignedString(ACCESS_SECRET)
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
	td.RefreshToken, err = rt.SignedString(REFRESH_SECRET)
	if err != nil {
		return nil, err
	}
	return td, nil
}

func verifyToken(tokenString string, secret []byte) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
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
