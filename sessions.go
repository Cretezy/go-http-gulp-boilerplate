package main

import (
	"os"

	"github.com/gorilla/sessions"
	"strconv"
)

type SessionManager struct {
	cookieStore *sessions.CookieStore
}

func NewSessionManager() *SessionManager {
	cookieSecret := os.Getenv("COOKIE_SECRET")
	if cookieSecret == "" {
		panic("COOKIE_SECRET not set")
	}

	secureString := os.Getenv("COOKIE_SECURE")
	if cookieSecret == "" {
		panic("COOKIE_SECURE not set")
	}

	secure, err := strconv.ParseBool(secureString)
	if err != nil {
		panic("COOKIE_SECURE invalid (must be true/false)")
	}

	cookieStore := sessions.NewCookieStore([]byte(cookieSecret))

	cookieStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		Secure:   secure,
		HttpOnly: true,
	}

	return &SessionManager{
		cookieStore,
	}
}
