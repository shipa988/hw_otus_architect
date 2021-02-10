package generator

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/shipa988/hw_otus_architect/internal/data/config"
	"github.com/shipa988/hw_otus_architect/internal/data/controller/log"
	"github.com/shipa988/hw_otus_architect/internal/data/repository/mysql"
	"github.com/shipa988/hw_otus_architect/internal/domain/usecase"
	"github.com/spf13/viper"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syreclabs.com/go/faker"
	"time"
)

type insertUser struct {
	Num      int
	Login    string
	Pass     string
	Name     string
	SurName  string
	Age      string
	Gen      string
	Interest string
	City     string
}

var signup int32
var saveprofile int32
var logout int32

func Generate(gencount int) {
	log.InitWithStdout("debug", "generator", "dev")
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	viper.SetConfigFile("c:\\Users\\redse\\go\\src\\github.com\\shipa988\\hw_otus_architect\\config\\network-prod.yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
	cfg := &config.Config{}
	err := viper.Unmarshal(cfg)
	if err != nil {
		log.Fatal(err)
	}
	repo := mysql.NewMySqlRepo()
	err = repo.Connect(ctx, cfg.DB)
	if err != nil {
		log.Fatal(err)
	}
	core := usecase.NewInteractor(repo, repo, repo, 15)
	uchan := make(chan insertUser)
	wg := &sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		time.Sleep(time.Millisecond*200)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for u := range uchan {
				log.Println("insert user:%v", u.Num)
				at, _, err := core.SignUp(u.Login, u.Name, u.Pass)
				if err != nil {
					log.Println(err)
					continue
				}
				atomic.AddInt32(&signup,1)
				token, err := base64.StdEncoding.DecodeString(at)
				if err != nil {
					log.Println(err)
					continue
				}
				sesid, id, err := core.VerifyUser(string(token), "at")
				if err != nil {
					log.Println(err)
					continue
				}
				userID, err := strconv.ParseUint(id, 10, 64)
				if err != nil {
					log.Println(err)
					continue
				}
				err = core.SaveMyProfile(userID, u.Name, u.SurName, u.Age, u.Gen, u.Interest, u.City)
				if err != nil {
					log.Println(err)
					continue
				}
				atomic.AddInt32(&saveprofile,1)
				err = core.Logout(userID, sesid)
				if err != nil {
					log.Println(err)
					continue
				}
				atomic.AddInt32(&logout,1)

			}

		}()
	}
	for i := 0; i < gencount; i++ {
		log.Println("generate user:%v", i)
		login :=  faker.Internet().UserName()+"_"+faker.RandomString(4)
		if len(login)>20{
			login=login[:20]
		}
			pass := faker.Internet().Password(5, 10)
		firstname := faker.Name().FirstName()
		lastname := faker.Name().LastName()
		age := strconv.Itoa(faker.RandomInt(12, 100))
		gender := faker.RandomChoice([]string{"male", "female", "other"})
		interests := faker.Lorem().String()
		city := faker.Address().City()
		uchan <- insertUser{
			Num:      i,
			Login:    login,
			Pass:     pass,
			Name:     firstname,
			SurName:  lastname,
			Age:      age,
			Gen:      gender,
			Interest: interests,
			City:     city,
		}
	}
	close(uchan)
	wg.Wait()
	fmt.Println("registred:",atomic.LoadInt32(&signup))
	fmt.Println("save profile:",atomic.LoadInt32(&saveprofile))
	fmt.Println("logout:",atomic.LoadInt32(&logout))
}
