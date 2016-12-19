package main

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/crypto/bcrypt"
)

// GetAuth Get auth page route
func GetAuth(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	templateManager.RenderView(w, r, "auth", nil, nil)
}

// PostAuth Post auth request route
func PostAuth(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	r.ParseForm()
	authType := r.Form.Get("type")
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	errors := []string{}

	if len(username) < 3 {
		errors = append(errors, "Username must be over 3 charactesr")
	} else if len(username) > 32 {
		errors = append(errors, "Username must be under 32 charactesr")
	}

	if len(password) < 8 {
		errors = append(errors, "Password must be over 8 charactesr")
	} else if len(password) > 256 {
		errors = append(errors, "Username must be under 256 charactesr")
	}

	if len([]string{}) > 0 {
		templateManager.RenderView(w, r, "auth", []string{}, nil)
	}

	// Get user
	user := &User{
		Username: username,
	}
	var count int
	db.Where(user).First(user).Count(&count)

	if authType == "login" {
		// Check user exist
		if count == 0 {
			// User doesn't exist
			templateManager.RenderView(w, r, "auth", []string{"Wrong username"}, nil)
			return
		}

		// Comparing the password with the hash
		err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			// Password does not match
			templateManager.RenderView(w, r, "auth", []string{"Wrong password"}, nil)
			return
		}
	} else if authType == "register" {
		// Check user doesn't exist
		if count != 0 {
			templateManager.RenderView(w, r, "auth", []string{"Username exist"}, nil)
			return
		}
		// Make new user
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}
		user = &User{
			Username: username,
			Password: string(hashedPassword),
		}
		db.Create(user)
	} else {
		templateManager.RenderView(w, r, "auth", nil, nil)
		return
	}


	// Create token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		panic(err)
	}
	token := fmt.Sprintf("%X", tokenBytes)

	session := &Session{
		UserID:       user.ID,
		Token:        token,
		LoginTime:    time.Now(),
		LastSeenTime: time.Now(),
	}

	db.Create(session)

	sessionStore, err := sessionManager.cookieStore.Get(r, "auth")
	if err != nil {
		panic(err)
	}

	sessionStore.Values["token"] = token
	sessionStore.Save(r, w)

	// Redirect to home
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func GetLogout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	sessionStore, err := sessionManager.cookieStore.Get(r, "auth")
	if err != nil {
		panic(err)
	}

	delete(sessionStore.Values, "token")
	sessionStore.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
