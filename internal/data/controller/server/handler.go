package server

import (
	"github.com/pkg/errors"
	"github.com/shipa988/hw_otus_architect/internal/data/controller/log"
	"html/template"
	"net/http"
	"path"
)

var (
	logintemplate = template.Must(template.ParseFiles(path.Join("web", "login.html")))
	friendstemplate = template.Must(template.ParseFiles(path.Join("web", "friends.html")))
	_404template  = template.Must(template.ParseFiles(path.Join("web", "404.html")))
)

func (s *HTTPServer) faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/ico/favicon.ico")
}

func (s *HTTPServer) _404Handler(w http.ResponseWriter, r *http.Request) {
	if e := _404template.Execute(w, ""); e != nil {
		log.Error(errors.Wrap(e, "_404 page error compose page"))
	}
}

func (s *HTTPServer) mainPage(w http.ResponseWriter, r *http.Request) {
	//if GetUserID(r.Context())!=""{
	s.httpAnswer(w, "main page", http.StatusOK)
	//	return
	//}
	//http.Redirect(w,r,"/login",302)
}

func (s *HTTPServer) loginPage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if e := logintemplate.Execute(w, struct {
			Err bool
		}{Err: false}); e != nil {
			log.Error(e)
			s.httpError(w, ErrServer, 500)
		}
	case http.MethodPost:
		r.ParseForm()
		user := r.FormValue("user")
		pass := r.FormValue("pass")
		err := s.networkcore.Login(user, pass)
		if err != nil {
			log.Error(errors.Wrapf(err, "can't authorize %v", user))
			if e := logintemplate.Execute(w, struct {
				Err bool
			}{Err: true}); e != nil {
				s.httpError(w, ErrServer, 500)
			}
			return
		}
		http.Redirect(w, r, "/", 302)
	default:
		s.httpError(w, "unknown method", http.StatusBadRequest)
		return
	}
}

func (s *HTTPServer) friends(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if e := logintemplate.Execute(w, struct {
			Err bool
		}{Err: false}); e != nil {
			log.Error(e)
			s.httpError(w, ErrServer, 500)
		}
	case http.MethodPost:
		r.ParseForm()
		user := r.FormValue("user")
		pass := r.FormValue("pass")
		err := s.networkcore.Login(user, pass)
		if err != nil {
			log.Error(errors.Wrapf(err, "can't authorize %v", user))
			if e := logintemplate.Execute(w, struct {
				Err bool
			}{Err: true}); e != nil {
				s.httpError(w, ErrServer, 500)
			}
			return
		}
		http.Redirect(w, r, "/", 302)
	default:
		s.httpError(w, "unknown method", http.StatusBadRequest)
		return
	}
}

func (s *HTTPServer) profile(w http.ResponseWriter, r *http.Request) {
	s.httpAnswer(w, "profile page", http.StatusOK)
}
