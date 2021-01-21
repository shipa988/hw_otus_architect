package server

import (
	"context"
	"net/http"
)

const UserID = contextKey("UserID")

type contextKey string

func GetUserID(ctx context.Context) (userID string) {
	if ctx == nil {
		return
	}
	userID, _ = ctx.Value(UserID).(string)
	return
}

func SetUserID(ctx context.Context, userID string) context.Context {
	if len(GetUserID(ctx)) == 0 {
		return context.WithValue(ctx, UserID, userID)
	}
	return ctx
}

type WrapResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *WrapResponseWriter) WriteHeader(status int) {
	w.ResponseWriter.WriteHeader(status)
	w.status = status
}


