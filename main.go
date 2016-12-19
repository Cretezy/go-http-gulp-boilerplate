package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/context"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
)

var sessionManager *SessionManager
var templateManager *TemplateManager
var db *gorm.DB

func GetHome(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if context.Get(r, "loggedIn").(bool) {
		templateManager.RenderView(w, r, "home.auth", nil, nil)
	} else {
		templateManager.RenderView(w, r, "home.guest", nil, nil)
	}
}

func main() {
	// Load environment variables from .env
	godotenv.Load()

	sessionManager = NewSessionManager()
	templateManager = NewTemplateManager()

	// Create database
	var err error
	db, err = gorm.Open("sqlite3", "test.db")
	if err != nil {
		log.Fatal("failed to connect database", err)
	}
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&User{}, &Session{}, )

	// Setup routes
	router := httprouter.New()
	router.GET("/", GetHome)
	router.GET("/auth", MustBeNotAuthed(GetAuth))
	router.POST("/auth", MustBeNotAuthed(PostAuth))
	router.GET("/auth/logout", GetLogout)

	router.GET("/account", MustBeAuthed(GetAccount))

	router.ServeFiles("/static/*filepath", &assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, AssetInfo: AssetInfo, Prefix: "public"})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start router
	fmt.Println("Starting base on port " + port)
	log.Fatal(http.ListenAndServe(":" + port, context.ClearHandler(AuthMiddleware(router))))
}

// AuthMiddleware Attachs user to request
func AuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var loggedIn bool
		sessionStore, err := sessionManager.cookieStore.Get(r, "auth")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, ok := sessionStore.Values["token"]; ok {
			token := sessionStore.Values["token"].(string)
			user := User{}
			session := &Session{Token: token}

			db.Where(session).First(session)
			db.Model(session).Related(&user)
			if user != (User{}) {
				session.LastSeenTime = time.Now()
				go db.Save(session)
				context.Set(r, "user", &user)
				loggedIn = true
			}
		}
		context.Set(r, "loggedIn", loggedIn)

		h.ServeHTTP(w, r)
	})
}

func MustBeAuthed(h httprouter.Handle) httprouter.Handle {
	return httprouter.Handle(func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		if context.Get(r, "loggedIn").(bool) {
			h(w, r, params)
		} else {
			http.Redirect(w, r, "/auth", http.StatusSeeOther)
		}
	})
}

func MustBeNotAuthed(h httprouter.Handle) httprouter.Handle {
	return httprouter.Handle(func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		if context.Get(r, "loggedIn").(bool) {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			h(w, r, params)
		}
	})
}
