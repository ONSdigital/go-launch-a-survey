package main // import "github.com/ONSdigital/go-launch-a-survey"

import (
	"fmt"

	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"

	"github.com/ONSdigital/go-launch-a-survey/authentication"
	"github.com/ONSdigital/go-launch-a-survey/settings"
	"github.com/ONSdigital/go-launch-a-survey/surveys"
)

func serveTemplate(templateName string, data interface{}, w http.ResponseWriter, r *http.Request) {
	lp := filepath.Join("templates", "layout.html")
	fp := filepath.Join("templates", filepath.Clean(templateName))

	// Return a 404 if the template doesn't exist or is directory
	info, err := os.Stat(fp)
	if err != nil && (os.IsNotExist(err) || info.IsDir()) {
		fmt.Println("Cannot find: " + fp)
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles(lp, fp)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

type page struct {
	Schemas []surveys.LauncherSchema
}

func getLaunchHandler(w http.ResponseWriter, r *http.Request) {
	p := page{Schemas: surveys.GetAvailableSchemas()}
	serveTemplate("launch.html", p, w, r)
}

func postLaunchHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, fmt.Sprintf("POST. r.ParseForm() err: %v", err), 500)
		return
	}

	token, tokenErr := authentication.ConvertPostToToken(r.PostForm)
	if tokenErr != nil {
		http.Error(w, fmt.Sprintf("ConvertPostToToken failed err: %v", tokenErr), 500)
		return
	}

	launchAction := r.PostForm.Get("action_launch")
	flushAction := r.PostForm.Get("action_flush")
	log.Println("Request: " + r.PostForm.Encode())

	hostURL := settings.Get("SURVEY_RUNNER_URL")

	if flushAction != "" {
		http.Redirect(w, r, hostURL+"/flush?token="+token, 307)
	} else if launchAction != "" {
		http.Redirect(w, r, hostURL+"/session?token="+token, 301)
	} else {
		http.Error(w, fmt.Sprintf("Invalid Action"), 500)
	}
}

func main() {
	r := mux.NewRouter()

	// Launch handlers
	r.HandleFunc("/", getLaunchHandler).Methods("GET")
	r.HandleFunc("/", postLaunchHandler).Methods("POST")

	// Serve static assets
	staticFs := http.FileServer(http.Dir("static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticFs))

	// Bind to a port and pass our router in
	hostname := settings.Get("GO_LAUNCH_A_SURVEY_LISTEN_HOST") + ":" + settings.Get("GO_LAUNCH_A_SURVEY_LISTEN_PORT")

	log.Println("Listening on " + hostname)
	log.Fatal(http.ListenAndServe(hostname, r))
}
