package surveys

import (
	"encoding/json"
	"log"
		"regexp"

	"github.com/AreaHQ/jsonhal"
	"github.com/ONSdigital/go-launch-a-survey/settings"
	"fmt"
	"io/ioutil"
	"strings"
	"sort"
	"github.com/ONSdigital/go-launch-a-survey/clients"
)

// LauncherSchema is a representation of a schema in the Launcher
type LauncherSchema struct {
	Name     string
	EqID     string
	FormType string
	URL      string
}

// LauncherSchemas is a separation of Test and Live schemas
type LauncherSchemas struct {
	Business []LauncherSchema
	Census   []LauncherSchema
	Social   []LauncherSchema
	Test     []LauncherSchema
	Other    []LauncherSchema
}

// RegisterResponse is the response from the eq-survey-register request
type RegisterResponse struct {
	jsonhal.Hal
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

// LauncherSchemaFromURL creates a LauncherSchema record from a url
func LauncherSchemaFromURL(url string) LauncherSchema {
	return LauncherSchema{
		URL: url,
	}
}

// GetAvailableSchemas Gets the list of static schemas an joins them with any schemas from the eq-survey-register if defined
func GetAvailableSchemas() LauncherSchemas {
	schemaList := LauncherSchemas{}

	runnerSchemas := getAvailableSchemasFromRunner()

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

	schemaList.Other = getAvailableSchemasFromRegister()

	sort.Sort(ByFilename(schemaList.Business))
	sort.Sort(ByFilename(schemaList.Census))
	sort.Sort(ByFilename(schemaList.Social))
	sort.Sort(ByFilename(schemaList.Test))

	return schemaList
}

// ByFilename implements sort.Interface based on the Name field.
type ByFilename []LauncherSchema

func (a ByFilename) Len() int           { return len(a) }
func (a ByFilename) Less(i, j int) bool { return a[i].Name < a[j].Name }
func (a ByFilename) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func getAvailableSchemasFromRegister() []LauncherSchema {

	schemaList := []LauncherSchema{}

	if settings.Get("SURVEY_REGISTER_URL") != "" {
		resp, err := clients.GetHTTPClient().Get(settings.Get("SURVEY_REGISTER_URL"))
		if err != nil {
			log.Fatal("Do: ", err)
			return []LauncherSchema{}
		}

		responseBody, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return schemaList
		}

		var registerResponse RegisterResponse
		if err := json.Unmarshal(responseBody, &registerResponse); err != nil {
			log.Print(err)
			return schemaList
		}

		var schemas Schemas

		schemasJSON, _ := json.Marshal(registerResponse.Embedded["schemas"])

		if err := json.Unmarshal(schemasJSON, &schemas); err != nil {
			log.Println(err)
		}

		for _, schema := range schemas {
			url := schema.Links["self"]
			EqID, formType := extractEqIDFormType(schema.Name)
			schemaList = append(schemaList, LauncherSchema{
				Name:     schema.Name,
				URL:      url.Href,
				EqID:     EqID,
				FormType: formType,
			})
		}
	}

	return schemaList
}

func getAvailableSchemasFromRunner() []LauncherSchema {

	schemaList := []LauncherSchema{}

	hostURL := settings.Get("SURVEY_RUNNER_SCHEMA_URL")

	log.Printf("Survey Runner Schema URL: %s", hostURL)

	url := fmt.Sprintf("%s/schemas", hostURL)

	resp, err := clients.GetHTTPClient().Get(url)
	if err != nil {
		return []LauncherSchema{}
	}

	if resp.StatusCode != 200 {
		return []LauncherSchema{}
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return []LauncherSchema{}
	}

	var schemaListResponse []string

	if err := json.Unmarshal(responseBody, &schemaListResponse); err != nil {
		log.Print(err)
		return []LauncherSchema{}
	}

	for _, schema := range schemaListResponse {
		schemaList = append(schemaList, LauncherSchemaFromFilename(schema))
	}

	return schemaList
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
	for _, survey := range availableSchemas.Other {
		if survey.Name == name {
			return survey
		}
	}
	panic("Survey not found")
}
