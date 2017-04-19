package main // import "github.com/collisdigital/go-launch-a-survey"

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"

	"github.com/collisdigital/go-launch-a-survey/authentication"
	"github.com/collisdigital/go-launch-a-survey/settings"
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

type Page struct {
	Schemas []string
}

func getAvailableSchemas() []string {
	// TODO: Replace with something dynamic
	return []string{
		"0_rogue_one.json",
		"0_star_wars.json",
		"1_0001.json",
		"1_0005.json",
		"1_0102.json",
		"1_0112.json",
		"1_0203.json",
		"1_0205.json",
		"1_0213.json",
		"1_0215.json",
		"2_0001.json",
		"census_communal.json",
		"census_household.json",
		"census_individual.json",
		"multiple_answers.json",
		"test_checkbox.json",
		"test_conditional_routing.json",
		"test_currency.json",
		"test_dates.json",
		"test_final_confirmation.json",
		"test_household_question.json",
		"test_interstitial_page.json",
		"test_language.json",
		"test_language_cy.json",
		"test_markup.json",
		"test_metadata_routing.json",
		"test_navigation.json",
		"test_navigation_confirmation.json",
		"test_percentage.json",
		"test_question_guidance.json",
		"test_radio.json",
		"test_radio_checkbox_descriptions.json",
		"test_relationship_household.json",
		"test_repeating_and_conditional_routing.json",
		"test_repeating_household.json",
		"test_routing_on_multiple_select.json",
		"test_skip_condition.json",
		"test_skip_condition_block.json",
		"test_skip_condition_group.json",
		"test_textarea.json",
		"test_textfield.json",
		"test_timeout.json",
		"test_total_breakdown.json"}
}

func getLaunchHandler(w http.ResponseWriter, r *http.Request) {
	page := Page{Schemas: getAvailableSchemas()}
	serveTemplate("launch.html", page, w, r)
}

func postLaunchHandler(w http.ResponseWriter, r *http.Request) {
	var token string
	var err error

	if err = r.ParseForm(); err != nil {
		fmt.Fprintf(w, "POST. ParseForm() err: %v\n", err)
		return
	}

	if token, err = authentication.ConvertPostToToken(r.PostForm); err == nil {
		hostUrl := settings.GetSetting("SURVEY_RUNNER_URL")
		http.Redirect(w, r, hostUrl+"/session?token="+token, 301)
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

	// Assign mux to handle all URLs
	http.Handle("/", r)

	// Bind to a port and pass our router in
	hostname := settings.GetSetting("GO_LAUNCH_A_SURVEY_LISTEN_HOST") + ":" + settings.GetSetting("GO_LAUNCH_A_SURVEY_LISTEN_PORT")

	log.Println("Listening on " + hostname)
	log.Fatal(http.ListenAndServe(hostname, nil))
}
