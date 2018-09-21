package main // import "github.com/ONSdigital/go-launch-a-survey"

import (
	"fmt"

	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"html"

	"github.com/ONSdigital/go-launch-a-survey/authentication"
	"github.com/ONSdigital/go-launch-a-survey/settings"
	"github.com/ONSdigital/go-launch-a-survey/surveys"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"gopkg.in/square/go-jose.v2/json"
)

func serveTemplate(templateName string, data interface{}, w http.ResponseWriter, r *http.Request) {
	lp := filepath.Join("templates", "layout.html")
	fp := filepath.Join("templates", filepath.Clean(templateName))

	// Return a 404 if the template doesn't exist or is directory
	info, err := os.Stat(fp)
	if err != nil && (os.IsNotExist(err) || info.IsDir()) {
		log.Println("Cannot find: " + fp)
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
	Schemas           surveys.LauncherSchemas
	AccountServiceURL string
}

func getStatusPage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func getLaunchHandler(w http.ResponseWriter, r *http.Request) {
	p := page{
		Schemas:           surveys.GetAvailableSchemas(),
		AccountServiceURL: getAccountServiceURL(r),
	}
	serveTemplate("launch.html", p, w, r)
}

func postLaunchHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, fmt.Sprintf("POST. r.ParseForm() err: %v", err), 500)
		return
	}
	redirectURL(w, r)
}

func getMetadataHandler(w http.ResponseWriter, r *http.Request) {
	schema := r.URL.Query().Get("schema")
	launcherSchema := surveys.FindSurveyByName(schema)

	metadata, err := authentication.GetRequiredMetadata(launcherSchema)

	if err != "" {
		http.Error(w, fmt.Sprintf("GetRequiredMetadata err: %v", err), 500)
		return
	}

	metadataJSON, _ := json.Marshal(metadata)

	w.Write([]byte(metadataJSON))

	return
}

func getAccountServiceURL(r *http.Request) string {
	forwardedProtocol := r.Header.Get("X-Forwarded-Proto")

	requestProtocol := "http"

	if forwardedProtocol != "" {
		requestProtocol = forwardedProtocol
	}

	return fmt.Sprintf("%s://%s",
		requestProtocol,
		html.EscapeString(r.Host))
}

func redirectURL(w http.ResponseWriter, r *http.Request) {
	hostURL := settings.Get("SURVEY_RUNNER_URL")

	token, err := authentication.GenerateTokenFromPost(r.PostForm)
	if err != "" {
		http.Error(w, err, 500)
		return
	}

	launchAction := r.PostForm.Get("action_launch")
	flushAction := r.PostForm.Get("action_flush")
	log.Println("Request: " + r.PostForm.Encode())

	if flushAction != "" {
		http.Redirect(w, r, hostURL+"/flush?token="+token, 307)
	} else if launchAction != "" {
		http.Redirect(w, r, hostURL+"/session?token="+token, 301)
	} else {
		http.Error(w, fmt.Sprintf("Invalid Action"), 500)
	}
}

func quickLauncherHandler(w http.ResponseWriter, r *http.Request) {
	hostURL := settings.Get("SURVEY_RUNNER_URL")
	accountServiceURL := getAccountServiceURL(r)
	urlValues := r.URL.Query()
	surveyURL := urlValues.Get("url")
	log.Println("Quick launch request received", surveyURL)

	urlValues.Add("ru_ref", authentication.GetDefaultValues()["ru_ref"])
	urlValues.Add("collection_exercise_sid", uuid.NewV4().String())
	urlValues.Add("case_id", uuid.NewV4().String())

	token, err := authentication.GenerateTokenFromDefaults(surveyURL, accountServiceURL, urlValues)
	if err != "" {
		http.Error(w, err, 500)
		return
	}

	if surveyURL != "" {
		http.Redirect(w, r, hostURL+"/session?token="+token, 302)
	} else {
		http.Error(w, fmt.Sprintf("Not Found"), 404)
	}
}

func main() {
	r := mux.NewRouter()

	// Launch handlers
	r.HandleFunc("/", getLaunchHandler).Methods("GET")
	r.HandleFunc("/", postLaunchHandler).Methods("POST")
	r.HandleFunc("/metadata", getMetadataHandler).Methods("GET")

	//Author Launcher with passed parameters in Url
	r.HandleFunc("/quick-launch", quickLauncherHandler).Methods("GET")

	// Status Page
	r.HandleFunc("/status", getStatusPage).Methods("GET")

	// Serve static assets
	staticFs := http.FileServer(http.Dir("static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticFs))

	// Bind to a port and pass our router in
	hostname := settings.Get("GO_LAUNCH_A_SURVEY_LISTEN_HOST") + ":" + settings.Get("GO_LAUNCH_A_SURVEY_LISTEN_PORT")

	log.Println("Listening on " + hostname)
	log.Fatal(http.ListenAndServe(hostname, r))
}
