package context

import (
	"context"
	"net/http"
)

type contextKey string

const UserIDKey contextKey = "sub"

func WithUserID(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), UserIDKey, userID)
	return r.WithContext(ctx)
}

func GetUserID(r *http.Request) string {
	if userID, ok := r.Context().Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}
