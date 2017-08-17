package surveys

import (
	"encoding/json"
	"github.com/ONSdigital/go-launch-a-survey/settings"
	"net/http"
	"log"
	"regexp"
	"github.com/AreaHQ/jsonhal"
)

// LauncherSchema is a representation of a schema in the Launcher
type LauncherSchema struct {
	Name     string
	EqID     string
	FormType string
	URL      string
}

// RegsiterResponse is the response from the eq-survey-register request
type RegsiterResponse struct {
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

// GetAvailableSchemas Gets the list of static schemas an joins them with any schemas from the eq-survey-register if defined
func GetAvailableSchemas() []LauncherSchema {
	schemaList := []LauncherSchema{
		LauncherSchemaFromFilename("0_star_wars.json"),
		LauncherSchemaFromFilename("1_0001.json"),
		LauncherSchemaFromFilename("1_0005.json"),
		LauncherSchemaFromFilename("1_0102.json"),
		LauncherSchemaFromFilename("1_0112.json"),
		LauncherSchemaFromFilename("1_0203.json"),
		LauncherSchemaFromFilename("1_0205.json"),
		LauncherSchemaFromFilename("1_0213.json"),
		LauncherSchemaFromFilename("1_0215.json"),
		LauncherSchemaFromFilename("2_0001.json"),
		LauncherSchemaFromFilename("census_communal.json"),
		LauncherSchemaFromFilename("census_household.json"),
		LauncherSchemaFromFilename("census_individual.json"),
		LauncherSchemaFromFilename("mci_0203.json"),
		LauncherSchemaFromFilename("mci_0213.json"),
		LauncherSchemaFromFilename("mci_0205.json"),
		LauncherSchemaFromFilename("multiple_answers.json"),
		LauncherSchemaFromFilename("rsi_0102.json"),
		LauncherSchemaFromFilename("rsi_0112.json"),
		LauncherSchemaFromFilename("test_big_list_naughty_strings.json"),
		LauncherSchemaFromFilename("test_checkbox.json"),
		LauncherSchemaFromFilename("test_conditional_dates.json"),
		LauncherSchemaFromFilename("test_conditional_routing.json"),
		LauncherSchemaFromFilename("test_currency.json"),
		LauncherSchemaFromFilename("test_dates.json"),
		LauncherSchemaFromFilename("test_final_confirmation.json"),
		LauncherSchemaFromFilename("test_household_question.json"),
		LauncherSchemaFromFilename("test_interstitial_page.json"),
		LauncherSchemaFromFilename("test_language.json"),
		LauncherSchemaFromFilename("test_language_cy.json"),
		LauncherSchemaFromFilename("test_markup.json"),
		LauncherSchemaFromFilename("test_metadata_routing.json"),
		LauncherSchemaFromFilename("test_navigation.json"),
		LauncherSchemaFromFilename("test_navigation_confirmation.json"),
		LauncherSchemaFromFilename("test_numbers.json"),
		LauncherSchemaFromFilename("test_percentage.json"),
		LauncherSchemaFromFilename("test_question_guidance.json"),
		LauncherSchemaFromFilename("test_radio.json"),
		LauncherSchemaFromFilename("test_radio_checkbox_descriptions.json"),
		LauncherSchemaFromFilename("test_relationship_household.json"),
		LauncherSchemaFromFilename("test_repeating_and_conditional_routing.json"),
		LauncherSchemaFromFilename("test_repeating_household.json"),
		LauncherSchemaFromFilename("test_routing_on_multiple_select.json"),
		LauncherSchemaFromFilename("test_skip_condition.json"),
		LauncherSchemaFromFilename("test_skip_condition_block.json"),
		LauncherSchemaFromFilename("test_skip_condition_group.json"),
		LauncherSchemaFromFilename("test_textarea.json"),
		LauncherSchemaFromFilename("test_textfield.json"),
		LauncherSchemaFromFilename("test_timeout.json"),
		LauncherSchemaFromFilename("test_total_breakdown.json"),
		LauncherSchemaFromFilename("test_unit_patterns.json"),
	}

	return append(schemaList, getAvailableSchemasFromRegister()...)
}

func getAvailableSchemasFromRegister() []LauncherSchema {

	schemaList := []LauncherSchema{}

	if settings.Get("SURVEY_REGISTER_URL") != "" {
		req, err := http.NewRequest("GET", settings.Get("SURVEY_REGISTER_URL"), nil)
		if err != nil {
			log.Fatal("NewRequest: ", err)
			return []LauncherSchema{}
		}
		client := &http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal("Do: ", err)
			return []LauncherSchema{}
		}

		defer resp.Body.Close()

		var registerResponse RegsiterResponse

		if err := json.NewDecoder(resp.Body).Decode(&registerResponse); err != nil {
			log.Println(err)
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
				Name: schema.Name,
				URL: url.Href,
				EqID: EqID,
				FormType: formType,
			})
		}
	}

	return schemaList
}

// FindSurveyByName Finds the schema in the list of available schemas
func FindSurveyByName(name string) LauncherSchema  {
	for _, survey := range GetAvailableSchemas() {
		if survey.Name == name {
			return survey
		}
	}
	panic("Survey not found")
}