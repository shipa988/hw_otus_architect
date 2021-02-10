package server

import (
	"github.com/pkg/errors"
	"github.com/shipa988/hw_otus_architect/internal/data/controller/log"
	"github.com/shipa988/hw_otus_architect/internal/domain/entity"
	"html/template"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

var (
	logintemplate   = template.Must(template.ParseFiles(path.Join("web", "login.html")))
	maintemplate    = template.Must(template.ParseFiles(path.Join("web", "main.html"), path.Join("web", "maintmpl.html")))
	othertemplate   = template.Must(template.ParseFiles(path.Join("web", "profile.html"), path.Join("web", "maintmpl.html")))
	signuptemplate  = template.Must(template.ParseFiles(path.Join("web", "signup.html")))
	friendstemplate = template.Must(template.ParseFiles(path.Join("web", "friends.html"), path.Join("web", "maintmpl.html")))
	searchtemplate  = template.Must(template.ParseFiles(path.Join("web", "search.html"), path.Join("web", "maintmpl.html")))
	_404template    = template.Must(template.ParseFiles(path.Join("web", "404.html")))
)

func (s *httpServer) faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/static/ico/favicon.ico")
}

func (s *httpServer) _404Handler(w http.ResponseWriter, r *http.Request) {
	if e := _404template.Execute(w, ""); e != nil {
		log.Error(errors.Wrap(e, "_404 page error compose page"))
	}
}

func (s *httpServer) messages(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/404", 302)
}

func (s *httpServer) mainPage(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(GetUserID(r.Context()), 10, 64)
	if err != nil {
		s.httpError(w, r, errors.Wrap(err, ErrServer).Error(), http.StatusInternalServerError)
		return
	}
	switch r.Method {
	case http.MethodGet:
		user, err := s.networkcore.GetMyProfile(userID)
		if err != nil {
			s.httpError(w, r, errors.Wrap(err, ErrServer).Error(), http.StatusInternalServerError)
			return
		}
		if e := maintemplate.Execute(w, struct {
			Err   bool
			Title string
			*entity.User
		}{Err: false, Title: "Your profile", User: &user}); e != nil {
			s.httpError(w, r, errors.Wrap(e, ErrServer).Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		r.ParseForm()
		name := r.FormValue("name")
		surname := r.FormValue("surname")
		city := r.FormValue("city")
		age := r.FormValue("age")
		gender := r.FormValue("gender")
		interests := r.FormValue("interests")
		err := s.networkcore.SaveMyProfile(userID, name, surname, age, gender, interests, city)
		if err != nil {
			log.Error(errors.Wrapf(err, "can't save profile for user %v", userID))
		}
		http.Redirect(w, r, "/", 302)
	default:
		s.httpError(w, r, ErrServer, http.StatusBadRequest)
	}
}

func (s *httpServer) loginPage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ref:=r.Header.Get("Referer")
		//check refresh token
		ctx, err := s.tokenVerify(r, "rt")
		if err != nil {
			log.Info(errors.Wrap(err, errAuth))
			if e := logintemplate.Execute(w, struct {
				Err bool
			}{Err: false}); e != nil {
				s.httpError(w, r, errors.Wrap(e, ErrServer).Error(), http.StatusInternalServerError)
			}
			return
		}
		userID, _ := strconv.ParseUint(GetUserID(ctx), 10, 64)

		accesstoken, refreshtoken, err := s.networkcore.SetTokenForUser(ctx, userID)
		if err != nil {
			log.Error(errors.Wrapf(err, "can't authorize user id '%v'", userID))
			if e := logintemplate.Execute(w, struct {
				Err bool
			}{Err: true}); e != nil {
				s.httpError(w, r, errors.Wrap(e, ErrServer).Error(), http.StatusInternalServerError)
			}
		}
		addCookie(w, "/login", "rt", refreshtoken, time.Hour*24*7)
		addCookie(w, "/", "at", accesstoken, time.Minute*15)
		redir,err:=url.Parse(ref)
		if err != nil {
			log.Error(errors.Wrapf(err, "can't perform true redirect fro user '%v'", userID))
			http.Redirect(w, r, "/", 302)
		}
		http.Redirect(w, r, redir.Path, 302)
	case http.MethodPost:
		r.ParseForm()
		login := r.FormValue("login")
		pass := r.FormValue("pass")
		accesstoken, refreshtoken, err := s.networkcore.Login(login, pass)
		if err != nil {
			log.Error(errors.Wrapf(err, "can't authorize %v", login))
			if e := logintemplate.Execute(w, struct {
				Err bool
			}{Err: true}); e != nil {
				s.httpError(w, r, errors.Wrap(e, ErrServer).Error(), http.StatusInternalServerError)
			}
		}
		addCookie(w, "/login", "rt", refreshtoken, time.Hour*24*7)
		addCookie(w, "/", "at", accesstoken, time.Minute*15)
		http.Redirect(w, r, "/", 302)
	default:
		s.httpError(w, r, ErrServer, http.StatusBadRequest)
	}
}

func (s *httpServer) signUpPage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if e := signuptemplate.Execute(w, struct {
			Err bool
		}{Err: false}); e != nil {
			s.httpError(w, r, errors.Wrap(e, ErrServer).Error(), http.StatusInternalServerError)
		}
	case http.MethodPost:
		r.ParseForm()
		login := r.FormValue("login")
		name := r.FormValue("name")
		pass := r.FormValue("pass")
		accesstoken, refreshtoken, err := s.networkcore.SignUp(login, name, pass)
		if err != nil {
			errstr := "register error"
			if nerr, ok := interface{}(errors.Cause(err)).(entity.ValidateError); ok {
				if nerr.IsWrongLenUserorPas() {
					errstr = "Login or Password length is invalid"
				}
				if nerr.IsLoginExist() {
					errstr = "Login alreay exist,PLease select another login."
				}
			}
			if e := signuptemplate.Execute(w, struct {
				Err string
			}{Err: errstr}); e != nil {
				s.httpError(w, r, errors.Wrap(e, ErrServer).Error(), http.StatusUnauthorized)
			}
			return
		}
		addCookie(w, "/login", "rt", refreshtoken, time.Hour*24*7)
		addCookie(w, "/", "at", accesstoken, time.Minute*15)
		http.Redirect(w, r, "/", 302)
	default:
		s.httpError(w, r, ErrServer, http.StatusBadRequest)
	}
}

func (s *httpServer) search(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if e := searchtemplate.Execute(w, struct {
			Err   bool
			Title string
		}{Err: false, Title: "People search"}); e != nil {
			s.httpError(w, r, errors.Wrap(e, ErrServer).Error(), http.StatusInternalServerError)
		}
	default:
		s.httpError(w, r, ErrServer, http.StatusBadRequest)
	}
}

func (s *httpServer) profile(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, err := strconv.ParseUint(r.FormValue("id"), 10, 64)
	if err != nil {
		s.httpError(w, r, errors.Wrap(err, ErrServer).Error(), http.StatusInternalServerError)
	}
	userID, err := strconv.ParseUint(GetUserID(r.Context()), 10, 64)
	if err != nil {
		s.httpError(w, r, errors.Wrap(err, ErrServer).Error(), http.StatusInternalServerError)
		return
	}
	switch r.Method {
	case http.MethodGet:
		profile, err := s.networkcore.GetUserProfile(userID, id)
		if err != nil {
			s.httpError(w, r, errors.Wrap(err, ErrServer).Error(), http.StatusInternalServerError)
			return
		}
		if e := othertemplate.Execute(w, struct {
			Err   bool
			Title string
			*entity.Profile
		}{Err: false, Title: "Profile of " + profile.Name + " " + profile.SurName, Profile: profile}); e != nil {
			s.httpError(w, r, errors.Wrap(e, ErrServer).Error(), http.StatusInternalServerError)
		}
	default:
		s.httpError(w, r, ErrServer, http.StatusBadRequest)
	}
}

func (s *httpServer) getpeople(w http.ResponseWriter, r *http.Request) {
	myid, err := strconv.ParseUint(GetUserID(r.Context()), 10, 64)
	if err != nil {
		s.httpError(w, r, errors.Wrap(err, ErrServer).Error(), http.StatusInternalServerError)
	}
	fr := r.FormValue("friends")
	if err != nil {
		s.httpError(w, r, errors.Wrap(err, ErrServer).Error(), http.StatusInternalServerError)
	}
	lastid, err := strconv.ParseUint(r.FormValue("lastid"), 10, 64)
	if err != nil {
		s.httpError(w, r, errors.Wrap(err, ErrServer).Error(), http.StatusInternalServerError)
	}
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		s.httpError(w, r, errors.Wrap(err, ErrServer).Error(), http.StatusInternalServerError)
	}
	if fr == "1" {
		friends, err := s.networkcore.GetFriends(myid, limit, lastid)
		if err != nil {
			s.httpError(w, r, errors.Wrap(err, ErrServer).Error(), http.StatusInternalServerError)
		}
		s.httpAnswer(w, friends, 200)
		return
	} else {
		name := r.FormValue("name")
		surname := r.FormValue("surname")
		friends, err := s.networkcore.GetPeople(myid, name, surname, limit, lastid)
		if err != nil {
			s.httpError(w, r, errors.Wrap(err, ErrServer).Error(), http.StatusInternalServerError)
		}
		s.httpAnswer(w, friends, 200)
		return
	}
	s.httpAnswer(w, "", 200)
}
func (s *httpServer) friends(w http.ResponseWriter, r *http.Request) {
	myid, err := strconv.ParseUint(GetUserID(r.Context()), 10, 64)
	if err != nil {
		s.httpError(w, r, errors.Wrap(err, ErrServer).Error(), http.StatusInternalServerError)
	}
	switch r.Method {
	case http.MethodGet:
		if e := friendstemplate.Execute(w, struct {
			Err   bool
			Title string
			Id    uint64
		}{Err: false, Title: "Friends", Id: myid}); e != nil {
			s.httpError(w, r, errors.Wrap(e, ErrServer).Error(), http.StatusInternalServerError)
		}
	default:
		s.httpError(w, r, ErrServer, http.StatusBadRequest)
	}
}

func (s *httpServer) logout(w http.ResponseWriter, r *http.Request) {
	myid, err := strconv.ParseUint(GetUserID(r.Context()), 10, 64)
	if err != nil {
		s.httpError(w, r, errors.Wrap(err, ErrServer).Error(), http.StatusInternalServerError)
	}
	if err := s.networkcore.Logout(myid, GetSessionUUID(r.Context())); err != nil {
		s.httpError(w, r, errors.Wrap(err, ErrServer).Error(), http.StatusInternalServerError)
		return
	}
	//addCookie(w, "/login", "rt", "", time.Hour*24*7)
	//addCookie(w, "/", "at", "", time.Minute*15)
	http.Redirect(w, r, "/", 302)
}

func (s *httpServer) subscribe(w http.ResponseWriter, r *http.Request) {
	myid, err := strconv.ParseUint(GetUserID(r.Context()), 10, 64)
	if err != nil {
		s.httpError(w, r, errors.Wrap(err, ErrServer).Error(), http.StatusInternalServerError)
	}
	r.ParseForm()
	id, err := strconv.ParseUint(r.FormValue("id"), 10, 64)
	if err != nil {
		s.httpError(w, r, errors.Wrap(err, ErrServer).Error(), http.StatusInternalServerError)
	}
	y := r.FormValue("yes")
	if err != nil {
		s.httpError(w, r, errors.Wrap(err, ErrServer).Error(), http.StatusInternalServerError)
	}
	if y == "1" {
		if err := s.networkcore.Subscribe(myid, id); err != nil {
			s.httpError(w, r, errors.Wrap(err, ErrServer).Error(), http.StatusInternalServerError)
		}
	} else {
		if err := s.networkcore.UnSubscribe(myid, id); err != nil {
			s.httpError(w, r, errors.Wrap(err, ErrServer).Error(), http.StatusInternalServerError)
		}
	}
	s.httpAnswer(w, "", 200)
}
func addCookie(w http.ResponseWriter, path, name, value string, ttl time.Duration) {
	expire := time.Now().Add(ttl)
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Expires:  expire,
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
}
