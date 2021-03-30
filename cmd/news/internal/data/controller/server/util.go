package server

import (
	"context"
	"net/http"
)

const (
	UserID = contextKey("UserID")
	SessionID = contextKey("SessionId")
)

type contextKey string


func GetUserID(ctx context.Context) (id string) {
	if ctx == nil {
		return
	}
	id, _ = ctx.Value(UserID).(string)
	return
}

func SetUserID(ctx context.Context, id string) context.Context {
	if len(GetUserID(ctx)) == 0 {
		return context.WithValue(ctx, UserID, id)
	}
	return ctx
}

func GetSessionUUID(ctx context.Context) (uuid string) {
	if ctx == nil {
		return
	}
	uuid, _ = ctx.Value(SessionID).(string)
	return
}

func SetSessionUUID(ctx context.Context, uuid string) context.Context {
	if len(GetSessionUUID(ctx)) == 0 {
		return context.WithValue(ctx, SessionID, uuid)
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


