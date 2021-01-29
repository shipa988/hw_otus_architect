package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/shipa988/hw_otus_architect/internal/data/controller/log"
	"github.com/shipa988/hw_otus_architect/internal/domain/usecase"
	"net/http"
	"path"
	"strconv"
	"time"
)

const (
	LayoutISO     = "2006-01-02 15:04:05"
	LayoutDateISO = "2006-01-02"
)
const (
	ErrID = "must be id in query"
	ErrServer = "Internal Server Error"
	errAuth = "can't authorize user"
	errTokenVerify = "can't verify token"
)
type HTTPServer struct {
	server *http.Server
	networkcore usecase.NetworkCore
}

func NewHttpServer(addr string, networkcore usecase.NetworkCore) *HTTPServer {
	server := &http.Server{
		Addr: addr,
	}
	return &HTTPServer{
		server: server,
		networkcore: networkcore,
	}
}

func (s *HTTPServer) Serve() error {
	log.Info("starting http server on address [%v]", s.server.Addr)

	privateMux := http.NewServeMux()
	privateMux.HandleFunc("/profile", s.profile)
	privateMux.HandleFunc("/friends", s.friends)
	privateMux.HandleFunc("/subscribe", s.subscribe)
	privateMux.HandleFunc("/getpeople", s.getpeople)
	privateMux.HandleFunc("/search", s.search)
	privateMux.HandleFunc("/logout", s.logout)
	privateMux.HandleFunc("/messages", s.messages)
	privateMux.HandleFunc("/", s.mainPage)
	privateHandler := s.authMiddleware(privateMux)

	publicMux := http.NewServeMux()
	publicMux.Handle("/", privateHandler)
	publicMux.HandleFunc("/login", s.loginPage)
	publicMux.HandleFunc("/signup", s.signUpPage)
	publicMux.HandleFunc("/404", s._404Handler)
	publicHandler := s.accessLogMiddleware(publicMux)

	staticMux := http.NewServeMux()
	staticMux.HandleFunc("/favicon.ico", s.faviconHandler)
	staticMux.Handle("/", publicHandler)
	fs := http.FileServer(http.Dir(path.Join("web", "static")))
	staticMux.Handle("/static/", http.StripPrefix("/static/", fs))

	siteHandler := s.corsHandler(staticMux)
	siteHandler = s.panicMiddleware(siteHandler)

	s.server.Handler = siteHandler

	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		return errors.Wrapf(err, "can't start listen address [%v]", s.server.Addr)
	}
	return nil
}


func (s *HTTPServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx,err := s.tokenVerify(r, "at")
		if err != nil {
			log.Error(errors.Wrap(err, errAuth))
			http.Redirect(w, r, "/login", 302)
			return
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *HTTPServer) tokenVerify(r *http.Request, tokenType string) (context.Context,error) {
	cookie, err := r.Cookie(tokenType)
	if err != nil {
		return nil,errors.Wrap(err, errTokenVerify)
	}
	token, err := base64.StdEncoding.DecodeString(cookie.Value)
	if err != nil {
		return nil,errors.Wrap(err, errTokenVerify)
	}
	sessionUuid,userID, err := s.networkcore.VerifyUser(string(token), tokenType)
	if err != nil {
		return nil,errors.Wrap(err, errTokenVerify)
	}
	return SetUserID(SetSessionUUID(r.Context(), sessionUuid), userID),nil
}

func (s *HTTPServer) StopServe() {
	ctx := context.Background()
	log.Info("stopping http server")
	defer log.Info("http server stopped")
	if s.server == nil {
		log.Error("http server is nil")
		return
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		log.Error("can't stop http server with error: %v", err)
	}
}

func (s *HTTPServer) accessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &WrapResponseWriter{ResponseWriter: w}
		next.ServeHTTP(rw, r)
		latency := time.Since(start)
		s.logRequest(r.RemoteAddr, start.Format(LayoutISO), r.Method, strconv.Itoa(rw.status), r.URL.Path, latency)
	})
}

func (s *HTTPServer) panicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				switch err.(type) {
				case error:
					log.Error(errors.Wrap(err.(error),ErrServer))
				default:
					log.Error(err,ErrServer)
				}
				http.Redirect(w, r, "/404", 302)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *HTTPServer) logRequest(remoteAddr, start, method, code, path string, latency time.Duration) {
	log.Info("%s [%s] %s %s %s [%dns]", remoteAddr, start, method, code, path, latency.Nanoseconds())
}

func (s *HTTPServer) corsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Add("Access-Control-Allow-Origin", "*")
		headers.Add("Vary", "Origin")
		headers.Add("Vary", "Access-Control-Request-Method")
		headers.Add("Vary", "Access-Control-Request-Headers")
		headers.Add("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, token")
		headers.Add("Access-Control-Allow-Methods", "GET, POST,OPTIONS")
		if r.Method == "OPTIONS" {
			s.httpAnswer(w, "", http.StatusOK)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func (s *HTTPServer) httpError(w http.ResponseWriter,  r *http.Request,error string, code int) {
	log.Error(error)
	http.Redirect(w, r, "/404", 302)
}

func (s *HTTPServer) httpAnswer(w http.ResponseWriter, msg interface{}, code int) {
	jmsg, err := json.Marshal(msg)
	if err != nil {
		code = http.StatusInternalServerError
	}
	w.WriteHeader(code)
	w.Write(jmsg) //nolint:errcheck
}





