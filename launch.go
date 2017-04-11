package main // import "github.com/collisdigital/go-launch-a-survey"

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"errors"
	"github.com/satori/go.uuid"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
	"time"

	"crypto/x509"
	"encoding/pem"
	"regexp"

	"github.com/collisdigital/go-launch-a-survey/settings"
)

type Page struct {
	Schemas []string
}

// TODO: Should this be combined with the below struct?
type PostFields struct {
	Jti                   string
	UserId                string
	Schema                string
	Exp                   string
	RuRef                 string
	RuName                string
	EqId                  string
	CollectionExerciseSid string
	PeriodId              string
	PeriodStr             string
	RefPStartDate         string
	RefPEndDate           string
	FormType              string
	ReturnBy              string
	TradAs                string
	EmploymentDate        string
	RegionCode            string
	LanguageCode          string
	TxId                  string
	VariantFlags          string
	Roles                 string
}

type EqClaims struct {
	jwt.Claims
	UserId                string `json:"user_id"`
	EqId                  string `json:"eq_id"`
	PeriodId              string `json:"period_id"`
	PeriodStr             string `json:"period_str"`
	CollectionExerciseSid string `json:"collection_exercise_sid"`
	RuRef                 string `json:"ru_ref"`
	RuName                string `json:"ru_name"`
	RefPStartDate         string `json:"ref_p_start_date"` // iso_8601_date
	RefPEndDate           string `json:"ref_p_end_date"`   // iso_8601_date
	FormType              string `json:"form_type"`
	ReturnBy              string `json:"return_by"`
	TradAs                string `json:"trad_as"`
	EmploymentDate        string `json:"employment_date"` // iso_8601_date
	RegionCode            string `json:"region_code"`
	LanguageCode          string `json:"language_code"`
	VariantFlags          string `json:"variant_flags"`
	Roles                 string `json:"roles"`
	TxId                  string `json:"tx_id"`
}

// TODO: should this move into a separate file?
// TODO: Document this, it returns an rsa.PublicKey
func loadEncryptionKey() (interface{}, error) {

	encryptionKeyPath := settings.GetSettings()["JWT_ENCRYPTION_KEY_PATH"]
	if keyData, err := ioutil.ReadFile(encryptionKeyPath); err == nil {
		block, _ := pem.Decode(keyData)
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err == nil {
			return pub, nil
		}
	}
	return nil, errors.New("Failed to load encryption key")
}

// TODO: should this move into a separate file?
// TODO: Document this, it returns an rsa.PrivateKey
// TODO: Add support for password protected private keys
func loadSigningKey() (interface{}, error) {
	signingKeyPath := settings.GetSettings()["JWT_SIGNING_KEY_PATH"]
	if keyData, err := ioutil.ReadFile(signingKeyPath); err == nil {
		block, _ := pem.Decode(keyData)
		priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err == nil {
			return priv, nil
		}
	}
	return nil, errors.New("Failed to load signing key")
}

func convertPostToToken(PostFields PostFields) (token string, err error) {
	log.Println("POST received...", PostFields)

	issued := time.Now()
	expires := issued.Add(time.Minute * 10)

	cl := EqClaims{
		Claims: jwt.Claims{
			IssuedAt: jwt.NewNumericDate(issued),
			Expiry:   jwt.NewNumericDate(expires),
			ID:       uuid.NewV4().String(),
		},
		UserId:                PostFields.UserId,
		EqId:                  PostFields.EqId,
		PeriodId:              PostFields.PeriodId,
		PeriodStr:             PostFields.PeriodStr,
		CollectionExerciseSid: PostFields.CollectionExerciseSid,
		RuRef:          PostFields.RuRef,
		RuName:         PostFields.RuName,
		RefPStartDate:  PostFields.RefPStartDate,
		RefPEndDate:    PostFields.RefPEndDate,
		FormType:       PostFields.FormType,
		ReturnBy:       PostFields.ReturnBy,
		TradAs:         PostFields.TradAs,
		EmploymentDate: PostFields.EmploymentDate,
		RegionCode:     PostFields.RegionCode,
		LanguageCode:   PostFields.LanguageCode,
		VariantFlags:   PostFields.VariantFlags,
		Roles:          PostFields.Roles,
		TxId:           uuid.NewV4().String(),
	}

	signingKey, err := loadSigningKey()
	if err != nil {
		fmt.Printf("Error loading signing key; err: %v", err)
		return
	}

	encryptionKey, err := loadEncryptionKey()
	if err != nil {
		fmt.Printf("Error loading encryption key; err: %v", err)
		return
	}

	opts := jose.SignerOptions{}
	opts.WithType("JWT")
	opts.WithHeader("kid", "EDCRRM")

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: signingKey}, &opts)
	if err != nil {
		fmt.Printf("Error creating JWT signer; err: %v", err)
		return
	}

	encryptor, err := jose.NewEncrypter(
		jose.A256GCM,
		jose.Recipient{Algorithm: jose.RSA_OAEP, Key: encryptionKey},
		(&jose.EncrypterOptions{}).WithType("JWT").WithContentType("JWT"))

	if err != nil {
		fmt.Printf("Error creating JWT encrypter; err: %v", err)
		return
	}

	token, err = jwt.SignedAndEncrypted(signer, encryptor).Claims(cl).CompactSerialize()

	if err != nil {
		fmt.Printf("Error signing and encrypting JWT; err: %v", err)
		return
	}

	fmt.Printf("Created signed/encrypted JWT: %v", token)
	return token, nil
}

func extractEqIdFormType(schema string) (eqId, formType string) {
	r := regexp.MustCompile(`^(?P<eq_id>[a-z0-9]+)_(?P<form_type>\w+)\.json`)
	match := r.FindStringSubmatch(schema)
	if match != nil {
		eqId = match[1]
		formType = match[2]
	}
	return
}

func rootHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		serveTemplate(w, r)
		return
	case "POST":
		var token string
		var err error

		if err = r.ParseForm(); err != nil {
			fmt.Fprintf(w, "POST. ParseForm() err: %v", err)
			return
		}

		fields := PostFields{}
		fields.UserId = r.PostForm.Get("user_id")
		fields.Exp = r.PostForm.Get("exp")
		fields.Schema = r.PostForm.Get("schema")
		fields.EqId, fields.FormType = extractEqIdFormType(fields.Schema)
		fields.PeriodStr = r.PostForm.Get("period_str")
		fields.PeriodId = r.PostForm.Get("period_id")
		fields.CollectionExerciseSid = r.PostForm.Get("collection_exercise_sid")
		fields.RefPStartDate = r.PostForm.Get("ref_p_start_date")
		fields.RefPEndDate = r.PostForm.Get("ref_p_end_date")
		fields.RuRef = r.PostForm.Get("ru_ref")
		fields.RuName = r.PostForm.Get("ru_name")
		fields.TradAs = r.PostForm.Get("trad_as")
		fields.ReturnBy = r.PostForm.Get("return_by")
		fields.EmploymentDate = r.PostForm.Get("employment_date")
		fields.RegionCode = r.PostForm.Get("region_code")
		fields.LanguageCode = r.PostForm.Get("language_code")

		// TODO: Add support for these fields
		//sexId := r.PostForm.Get("sexual_identity")
		//fields.VariantFlags = {"sexual_identity": sexual_identity}
		//fields.Roles = ['dumper']

		if token, err = convertPostToToken(fields); err == nil {
			hostUrl := settings.GetSettings()["SURVEY_RUNNER_URL"]
			http.Redirect(w, r, hostUrl+"/session?token="+token, 301)
		}

	default:
		fmt.Fprintf(w, "Only GET and POST methods are supported.")
	}
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

func serveTemplate(w http.ResponseWriter, r *http.Request) {
	lp := filepath.Join("templates", "layout.html")
	fp := filepath.Join("templates", filepath.Clean(r.URL.Path))

	// Return a 404 if the template doesn't exist or is directory
	info, err := os.Stat(fp)
	if err != nil && (os.IsNotExist(err) || info.IsDir()) {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles(lp, fp)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}

	page := Page{Schemas: getAvailableSchemas()}
	if err := tmpl.ExecuteTemplate(w, "layout", page); err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

func main() {
	// Serve static assets
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Serve templates
	http.HandleFunc("/", rootHandler)

	// Bind to a port and pass our router in
	hostPort := settings.GetSettings()["GO_LAUNCH_A_SURVEY_LISTEN_HOST"] + ":" + settings.GetSettings()["GO_LAUNCH_A_SURVEY_LISTEN_PORT"]

	log.Println("Listening on " + hostPort)
	log.Fatal(http.ListenAndServe(hostPort, nil))
}
