package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

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
		port = "3001"
	}

	fs := http.FileServer(http.Dir("assets"))
	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))

	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/doc", docHandler)
	mux.HandleFunc("/pdf", pdfHandler)
	http.ListenAndServe(":"+port, mux)
}
