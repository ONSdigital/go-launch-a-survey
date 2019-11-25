package surveys

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/AreaHQ/jsonhal"
	"github.com/ONSdigital/go-launch-a-survey/clients"
	"github.com/ONSdigital/go-launch-a-survey/settings"
)

// ReqVersionBodyParams is a representation of the body params for the request
type ReqVersionBodyParams struct {
	SurveyID      string
	FormType      string
	SurveyVersion string
}

// LauncherSchema is a representation of a schema in the Launcher
type LauncherSchema struct {
	Name       string
	EqID       string
	FormType   string
	URL        string
	BodyParams ReqVersionBodyParams
}

// LauncherSchemas is a separation of Test and Live schemas
type LauncherSchemas struct {
	Business []LauncherSchema
	Census   []LauncherSchema
	Social   []LauncherSchema
	Test     []LauncherSchema
	Register []LauncherSchema
	Other    []LauncherSchema
}

// RegisterResponse is the response from the eq-survey-register request
type RegisterResponse struct {
	FormType      string `json:"form_type"`
	LastPublished string `json:"lastPublished"`
	RegistryID    string `json:"registry_id"`
	EqID          string `json:"eq_id"`
	SurveyID      string `json:"survey_id"`
	SurveyVersion string `json:"survey_version"`
	Title         string `json:"title"`
}

// Schemas is a list of Schema
type Schemas []Schema

// Schema is an available schema
type Schema struct {
	jsonhal.Hal
	Name string `json:"name"`
}

var eqIDFormTypeRegex = regexp.MustCompile(`^(?P<eq_id>[a-z0-9]+)_(?P<form_type>\w+)`)

func extractEqIDFormType(schema string) (EqID, formType string) {
	match := eqIDFormTypeRegex.FindStringSubmatch(schema)
	if match != nil {
		EqID = match[1]
		formType = match[2]
	}
	return
}

// LauncherSchemaFromFilename creates a LauncherSchema record from a schema filename
func LauncherSchemaFromFilename(filename string) LauncherSchema {
	EqID, formType := extractEqIDFormType(filename)
	return LauncherSchema{
		Name:     filename,
		EqID:     EqID,
		FormType: formType,
	}
}

// GetAvailableSchemas Gets the list of static schemas an joins them with any schemas from the eq-survey-register if defined
func GetAvailableSchemas() LauncherSchemas {

	client := clients.GetHTTPClient()

	schemaList := LauncherSchemas{}

	runnerSchemas, err := getAvailableSchemasFromRunner(client)
	if err != nil {
		log.Printf(`WARN: Unexpected error whilst retrieving schemas from EQRunner; skipping repository`)
	} else {
		for _, launcherSchema := range runnerSchemas {
			if strings.HasPrefix(launcherSchema.Name, "test_") {
				schemaList.Test = append(schemaList.Test, launcherSchema)
			} else if strings.HasPrefix(launcherSchema.Name, "census_") {
				schemaList.Census = append(schemaList.Census, launcherSchema)
			} else if strings.HasPrefix(launcherSchema.Name, "lms_") {
				schemaList.Social = append(schemaList.Social, launcherSchema)
			} else {
				schemaList.Business = append(schemaList.Business, launcherSchema)
			}
		}
	}

	schemaList.Register, err = GetAvailableSchemasFromRegister(client)
	if err != nil {
		log.Print(err)
	}

	sort.Sort(ByFilename(schemaList.Business))
	sort.Sort(ByFilename(schemaList.Census))
	sort.Sort(ByFilename(schemaList.Social))
	sort.Sort(ByFilename(schemaList.Test))
	sort.Sort(ByFilename(schemaList.Register))
	sort.Sort(ByFilename(schemaList.Other))

	return schemaList
}

// ByFilename implements sort.Interface based on the Name field.
type ByFilename []LauncherSchema

func (a ByFilename) Len() int           { return len(a) }
func (a ByFilename) Less(i, j int) bool { return a[i].Name < a[j].Name }
func (a ByFilename) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// GetAvailableSchemasFromRegister Gets published questionnaires from register
func GetAvailableSchemasFromRegister(httpClient *http.Client) ([]LauncherSchema, error) {

	schemaList := []LauncherSchema{}

	registerURL := settings.Get("SURVEY_REGISTER_URL")

	url := fmt.Sprintf("%s/questionnaires/published", registerURL)

	if registerURL != "" {

		resp, err := httpClient.Get(url)
		if err != nil {
			return nil, fmt.Errorf("WARN: Failed to contact %s; skipping schema repository", url)
		}

		responseBody, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("WARN: Failed to read response from %s; skipping schema repository", url)
		}

		var questionnaires []RegisterResponse

		err = json.Unmarshal(responseBody, &questionnaires)
		if err != nil {
			return nil, fmt.Errorf("WARN: Failed to unmarshall response from %s; skipping schema repository", url)
		}

		for _, questionnaire := range questionnaires {

			t, err := time.Parse(time.RFC3339, questionnaire.LastPublished)
			if err != nil {
				log.Printf("WARN: Failed to parse LastPublished date; skipping questionnaire: %s_%s", questionnaire.SurveyID, questionnaire.FormType)
				continue
			}

			numOfVersions, err := strconv.Atoi(questionnaire.SurveyVersion)
			if err != nil {
				log.Printf("WARN: convert survey version string to int; skipping questionnaire: %s_%s", questionnaire.SurveyID, questionnaire.FormType)
				continue
			}

			for i := 1; i <= numOfVersions; i++ {

				schemaList = append(schemaList, LauncherSchema{
					Name:     fmt.Sprintf("%s_%s %s (v%d - %d/%d/%d)", questionnaire.SurveyID, questionnaire.FormType, questionnaire.Title, i, t.Day(), t.Month(), t.Year()),
					URL:      fmt.Sprintf("%s/questionnaires/version?survey_id=%s&form_type=%s&survey_version=%d", registerURL, questionnaire.SurveyID, questionnaire.FormType, i),
					EqID:     questionnaire.EqID,
					FormType: questionnaire.FormType,
				})
			}
		}
	}

	return schemaList, nil
}

func getAvailableSchemasFromRunner(httpClient *http.Client) ([]LauncherSchema, error) {

	schemaList := []LauncherSchema{}

	hostURL := settings.Get("SURVEY_RUNNER_SCHEMA_URL")

	url := fmt.Sprintf("%s/schemas", hostURL)

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to contact EQRunner at %s", url)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Contacted EQRunner but returned unexpected status code; expected 200 but got %d", resp.StatusCode)
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("Unexpected error whilst reading EQRunner response: %s", err)
	}

	var schemaListResponse []string

	if err := json.Unmarshal(responseBody, &schemaListResponse); err != nil {
		log.Print(err)
		return nil, fmt.Errorf("Unexpected error whilst unmarshaling EQRunner response: %s", err)
	}

	for _, schema := range schemaListResponse {
		schemaList = append(schemaList, LauncherSchemaFromFilename(schema))
	}

	return schemaList, nil
}

// FindSurveyByName Finds the schema in the list of available schemas
func FindSurveyByName(name string) LauncherSchema {
	availableSchemas := GetAvailableSchemas()

	for _, survey := range availableSchemas.Business {
		if survey.Name == name {
			return survey
		}
	}
	for _, survey := range availableSchemas.Census {
		if survey.Name == name {
			return survey
		}
	}
	for _, survey := range availableSchemas.Social {
		if survey.Name == name {
			return survey
		}
	}
	for _, survey := range availableSchemas.Test {
		if survey.Name == name {
			return survey
		}
	}
	for _, survey := range availableSchemas.Register {
		if survey.Name == name {
			return survey
		}
	}
	for _, survey := range availableSchemas.Other {
		if survey.Name == name {
			return survey
		}
	}
	panic("Survey not found")
}
