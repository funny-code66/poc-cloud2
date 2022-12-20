package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Scopes: OAuth 2.0 scopes provide a way to limit the amount of access that is granted to an access token.
var googleOauthConfig = &oauth2.Config{
	RedirectURL:  "http://127.0.0.1:8000/auth/google/callback/",
	ClientID:     "284873503032-3apt2hr48dn75oql589h6g1tui3f79ih.apps.googleusercontent.com",
	ClientSecret: "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",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

func oauthGoogleLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("----------------oauthGoogleLogin()")

	// Create oauthState cookie
	oauthState := generateStateOauthCookie(w)
	log.Println("----------------create oauthState cookie")
	/*
		AuthCodeURL receive state that is a token to protect the user from CSRF attacks. You must always provide a non-empty string and
		validate that it matches the the state query parameter on your redirect callback.
	*/
	u := googleOauthConfig.AuthCodeURL(oauthState)
	log.Println("----------------create AuthCodeURL")
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

func oauthGoogleCallback(w http.ResponseWriter, r *http.Request) {
	log.Println("----------------oauthGoogleCallback()")
	// Read oauthState from Cookie
	// oauthState, _ := r.Cookie("oauthstate")
	log.Println("----------------Read oauthState from Cookie")
	log.Println("----------------%s", r.FormValue("state"))

	// if r.FormValue("state") != oauthState.Value {
	// 	log.Println("----------------invalid oauth google state")
	// 	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	// 	return
	// }

	data, err := getUserDataFromGoogle(r.FormValue("code"))
	log.Println("----------------getUserDataFromGoogle()")
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// This time, you can see following response on website
	/* UserInfo: {
	   		"id": "******************",
				"email": "****************@gmail.com",
				"verified_email": true,
				"picture": "https://lh3.googleusercontent.com/a-/*********************************"
			}
	*/
	fmt.Fprintf(w, "UserInfo: %s\n", data)
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(20 * time.Minute)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}

func getUserDataFromGoogle(code string) ([]byte, error) {
	// Use code to get token and get user info from Google.

	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	response, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	}
	return contents, nil
}

type PageVariables struct {
	Date string
	Time string
}

type TxData struct {
	Title   string
	Content string
}

var tpl = template.Must(template.ParseFiles("index.html"))

func docHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("doc handler")
	now := time.Now()              // find the time right now
	HomePageVars := PageVariables{ //store the date and time in a struct
		Date: now.Format("02-01-2006"),
		Time: now.Format("15:04:05"),
	}
	log.Printf("get fetch: %v", r.Body)
	var t TxData
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&t)

	if err != nil {
		panic(err)
	}
	log.Printf("Title: %s", t.Title)
	log.Printf("Content: %s", t.Content)

	// w.Write([]byte("<h1>Hello World!</h1>"))
	tpl.Execute(w, HomePageVars)
}

func pdfHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("pdf handler")
	// now := time.Now() // find the time right now
	// HomePageVars := PageVariables{ //store the date and time in a struct
	// 	Date: now.Format("02-01-2006"),
	// 	Time: now.Format("15:04:05"),
	// }
	w.Write([]byte("<h1>Hello World!</h1>"))
	// tpl.Execute(w, HomePageVars)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	// now := time.Now() // find the time right now
	// HomePageVars := PageVariables{ //store the date and time in a struct
	// 	Date: now.Format("02-01-2006"),
	// 	Time: now.Format("15:04:05"),
	// }
	// w.Write([]byte("<h1>Hello World!</h1>"))

	tpl.Execute(w, nil)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	fs := http.FileServer(http.Dir("assets"))
	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// Root
	mux.HandleFunc("/", indexHandler)

	// OauthGoogle
	mux.HandleFunc("/auth/google/login", oauthGoogleLogin)
	mux.HandleFunc("/auth/google/callback/", oauthGoogleCallback)

	// FileGoogle
	mux.HandleFunc("/doc", docHandler)
	mux.HandleFunc("/pdf", pdfHandler)

	log.Printf("Starting HTTP Server. Listening at %q", port)
	if err := http.ListenAndServe(":"+port, mux); err != http.ErrServerClosed {
		log.Printf("%v", err)
	} else {
		log.Println("Server closed!")
	}
}
