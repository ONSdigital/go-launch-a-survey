package surveys

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/ONSdigital/go-launch-a-survey/settings"
)

func TestIfLauncherCanMakeCallToRegisterAPI(t *testing.T) {

	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(bytes.NewBufferString(`
			[
				{
					"registry_id": "b02f1331-57f3-4427-8182-c969dbed6414",
					"survey_id": "187",
					"form_type": "002",
					"title": "Ecommerce",
					"lastPublished": "2019-12-12T08:55:27.731Z",
					"survey_version": "1"
				}
			]`)),
			Header: make(http.Header),
		}
	})

	registerURL := settings.Get("SURVEY_REGISTER_URL")
	url := fmt.Sprintf("%s/published-questionnaires", registerURL)
	potentialErr := fmt.Errorf("WARN: Failed to contact %s; skipping schema repository", url)

	_, err := GetAvailableSchemasFromRegister(client)
	if err == potentialErr {
		t.Errorf("Failed to contact %s", url)
	}
	if err != nil {
		t.Errorf("Error %s recieved, expected nil", err)
	}
}

func TestIfLauncherCanBuildLaunchSchemaFromRegisterResponse(t *testing.T) {
	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(bytes.NewBufferString(`
			[
				{
					"registry_id": "b02f1331-57f3-4427-8182-c969dbed6414",
					"survey_id": "187",
					"form_type": "002",
					"title": "Ecommerce",
					"lastPublished": "2019-12-12T08:55:27.731Z",
					"survey_version": "1",
					"eq_id": "123-456-789"
				}
			]`)),
			Header: make(http.Header),
		}
	})

	registerURL := settings.Get("SURVEY_REGISTER_URL")

	var expectedLauncherSchemas []LauncherSchema

	expectedLauncherSchemas = append(expectedLauncherSchemas, LauncherSchema{
		Name:     "187_002 Ecommerce (v1 - 12/12/2019)",
		URL:      fmt.Sprintf(`%s/questionnaire/187/002/1`, registerURL),
		EqID:     "123-456-789",
		FormType: "002",
	})

	launcherSchema, err := GetAvailableSchemasFromRegister(client)

	if err != nil {
		t.Errorf("Error %s recieved, expected nil", err)
	}
	if launcherSchema[0] != expectedLauncherSchemas[0] {
		t.Errorf("Built launcherSchema incorrectly; expected %s but recieved %s", expectedLauncherSchemas, launcherSchema)
	}

}

func TestIfLauncherCanMakeCallToEqRunnerAPI(t *testing.T) {
	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 200,
			Body: ioutil.NopCloser(bytes.NewBufferString(`[
					"test_routing_answer_comparison.json",
					"test_titles_repeating_non_repeating_dependency.json",
					"census_communal.json",
					"mbs_0216.json"]`)),
			Header: make(http.Header),
		}
	})

	_, err := getAvailableSchemasFromRunner(client)

	surveyRunnerURL := settings.Get("SURVEY_RUNNER_URL")
	url := fmt.Sprintf("%s/published-questionnaires", surveyRunnerURL)
	potentialErr := fmt.Errorf("WARN: Failed to contact %s; skipping schema repository", url)
	if err == potentialErr {
		t.Errorf("Failed to contact %s", url)
	}
	if err != nil {
		t.Errorf("Error %s recieved, expected nil", err)
	}
}

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}
