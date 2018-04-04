package authentication

import (
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/ONSdigital/go-launch-a-survey/settings"
	"github.com/ONSdigital/go-launch-a-survey/surveys"
	"github.com/satori/go.uuid"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
	"io/ioutil"
	"net/url"
	"time"
	"net/http"
	"gopkg.in/square/go-jose.v2/json"

	"math/rand"
	"encoding/base64"
	"log"
	"bytes"
	"strings"
	"path"
)

// KeyLoadError describes an error that can occur during key loading
type KeyLoadError struct {
	// Op is the operation which caused the error, such as
	// "read", "parse" or "cast".
	Op string

	// Err is a description of the error that occurred during the operation.
	Err string
}

func (e *KeyLoadError) Error() string {
	if e == nil {
		return "<nil>"
	}
	return e.Op + ": " + e.Err
}

// PublicKeyResult is a wrapper for the public key and the kid that identifies it
type PublicKeyResult struct {
	key *rsa.PublicKey
	kid string
}

// PrivateKeyResult is a wrapper for the private key and the kid that identifies it
type PrivateKeyResult struct {
	key *rsa.PrivateKey
	kid string
}

func loadEncryptionKey() (*PublicKeyResult, *KeyLoadError) {
	encryptionKeyPath := settings.Get("JWT_ENCRYPTION_KEY_PATH")

	keyData, err := ioutil.ReadFile(encryptionKeyPath)
	if err != nil {
		return nil, &KeyLoadError{Op: "read", Err: "Failed to read encryption key from file: " + encryptionKeyPath}
	}

	block, _ := pem.Decode(keyData)
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, &KeyLoadError{Op: "parse", Err: "Failed to parse encryption key PEM"}
	}

	kid := fmt.Sprintf("%x", sha1.Sum(keyData))

	publicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, &KeyLoadError{Op: "cast", Err: "Failed to cast key to rsa.PublicKey"}
	}

	return &PublicKeyResult{publicKey, kid}, nil
}

func loadSigningKey() (*PrivateKeyResult, *KeyLoadError) {
	signingKeyPath := settings.Get("JWT_SIGNING_KEY_PATH")
	keyData, err := ioutil.ReadFile(signingKeyPath)
	if err != nil {
		return nil, &KeyLoadError{Op: "read", Err: "Failed to read signing key from file: " + signingKeyPath}
	}

	block, _ := pem.Decode(keyData)
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, &KeyLoadError{Op: "parse", Err: "Failed to parse signing key from PEM"}
	}

	PublicKey, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, &KeyLoadError{Op: "marshal", Err: "Failed to marshal public key"}
	}

	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: PublicKey,
	})
	kid := fmt.Sprintf("%x", sha1.Sum(pubBytes))

	return &PrivateKeyResult{privateKey, kid}, nil
}

// EqClaims is a representation of the set of values needed when generating a valid token
type EqClaims struct {
	jwt.Claims
	UserID                string       `json:"user_id"`
	EqID                  string       `json:"eq_id"`
	PeriodID              string       `json:"period_id"`
	PeriodStr             string       `json:"period_str"`
	CollectionExerciseSid string       `json:"collection_exercise_sid"`
	RuRef                 string       `json:"ru_ref"`
	RuName                string       `json:"ru_name"`
	RefPStartDate         string       `json:"ref_p_start_date"`         // iso_8601_date
	RefPEndDate           string       `json:"ref_p_end_date,omitempty"` // iso_8601_date
	FormType              string       `json:"form_type"`
	SurveyURL             string       `json:"survey_url,omitempty"`
	ReturnBy              string       `json:"return_by"`
	TradAs                string       `json:"trad_as,omitempty"`
	EmploymentDate        string       `json:"employment_date,omitempty"` // iso_8601_date
	RegionCode            string       `json:"region_code,omitempty"`
	LanguageCode          string       `json:"language_code,omitempty"`
	VariantFlags          variantFlags `json:"variant_flags,omitempty"`
	TxID                  string       `json:"tx_id,omitempty"`
	Roles                 []string     `json:"roles,omitempty"`
	CaseID                string       `json:"case_id,omitempty"`
	CaseRef               string       `json:"case_ref,omitempty"`
	AccountServiceURL     string       `json:"account_service_url,omitempty"`
}

type variantFlags struct {
	SexualIdentity bool `json:"sexual_identity,omitempty"`
}

// QuestionnaireSchema is a minimal representation of a questionnaire schema used for extracting the eq_id and form_type
type QuestionnaireSchema struct {
	EqID     string `json:"eq_id"`
	FormType string `json:"form_type"`
}

// Generates a random string of a defined length
func randomStringGen() (rs string) {
	size := 6 //change the length of the generated random string
	rb := make([]byte, size)
	_, err := rand.Read(rb)

	if err != nil {
		log.Println(err)
	}

	randomString := base64.URLEncoding.EncodeToString(rb)

	return randomString
}

func generateDefaultClaims(accountURL string) (claims EqClaims) {
	defaultClaims := EqClaims{
		UserID:                "UNKNOWN",
		PeriodID:              "201605",
		PeriodStr:             "May 2017",
		CollectionExerciseSid: randomStringGen(),
		RuRef:                 "12346789012A",
		RuName:                "ESSENTIAL ENTERPRISE LTD.",
		RefPStartDate:         "2016-05-01",
		RefPEndDate:           "2016-05-31",
		ReturnBy:              "2016-06-12",
		TradAs:                "ESSENTIAL ENTERPRISE LTD.",
		EmploymentDate:        "2016-06-10",
		RegionCode:            "GB-ENG",
		LanguageCode:          "en",
		TxID:                  uuid.NewV4().String(),
		CaseID:                uuid.NewV4().String(),
		CaseRef:               "1000000000000001",
		AccountServiceURL:     accountURL,
		VariantFlags: variantFlags{
			SexualIdentity: true,
		},
		Roles: []string{"dumper"},
	}
	return defaultClaims
}

func generateClaimsFromPost(postValues url.Values) (claims EqClaims) {
	postClaims := EqClaims{
		UserID:                postValues.Get("user_id"),
		PeriodID:              postValues.Get("period_id"),
		PeriodStr:             postValues.Get("period_str"),
		CollectionExerciseSid: postValues.Get("collection_exercise_sid"),
		RuRef:                 postValues.Get("ru_ref"),
		RuName:                postValues.Get("ru_name"),
		RefPStartDate:         postValues.Get("ref_p_start_date"),
		RefPEndDate:           postValues.Get("ref_p_end_date"),
		ReturnBy:              postValues.Get("return_by"),
		TradAs:                postValues.Get("trad_as"),
		EmploymentDate:        postValues.Get("employment_date"),
		RegionCode:            postValues.Get("region_code"),
		LanguageCode:          postValues.Get("language_code"),
		TxID:                  uuid.NewV4().String(),
		VariantFlags: variantFlags{
			SexualIdentity: postValues.Get("sexual_identity") == "true",
		},
		Roles:             []string{postValues.Get("roles")},
		CaseID:            uuid.NewV4().String(),
		CaseRef:           postValues.Get("case_ref"),
		AccountServiceURL: postValues.Get("account_url"),
	}

	return postClaims
}

// GenerateJwtClaims creates a jwtClaim needed to generate a token
func GenerateJwtClaims() (jwtClaims jwt.Claims) {
	issued := time.Now()
	expires := issued.Add(time.Minute * 10) // TODO: Support custom exp: r.PostForm.Get("exp")

	jwtClaims = jwt.Claims{
		IssuedAt: jwt.NewNumericDate(issued),
		Expiry:   jwt.NewNumericDate(expires),
		ID:       uuid.NewV4().String(),
	}

	return jwtClaims
}

func launcherSchemaFromURL(url string) (launcherSchema surveys.LauncherSchema, error string) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		return launcherSchema, fmt.Sprintf("Failed to load Schema from %s", url)
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	validationError := validateSchema(responseBody)
	if validationError != "" {
		return launcherSchema, validationError
	}

	var schema QuestionnaireSchema
	if err := json.Unmarshal(responseBody, &schema); err != nil {
		panic(err)
	}

	cacheBust := ""
	if !strings.Contains(url, "?") {
		cacheBust = "?bust=" + time.Now().Format("20060102150405")
	}

	launcherSchema = surveys.LauncherSchema{
		EqID:     schema.EqID,
		FormType: schema.FormType,
		URL:      url + cacheBust,
	}

	return launcherSchema, ""
}

func validateSchema(payload []byte) (error string) {
	if settings.Get("SCHEMA_VALIDATOR_URL") == "" {
		return ""
	}

	validateURL := path.Join(settings.Get("SCHEMA_VALIDATOR_URL"), "validate")
	resp, err := http.Post(validateURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err.Error()
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err.Error()
	}

	if resp.StatusCode != 200 {
		return string(responseBody)
	}

	return ""
}

func addSchemaToClaims(claims *EqClaims, LauncherSchema surveys.LauncherSchema) () {
	claims.EqID = LauncherSchema.EqID
	claims.FormType = LauncherSchema.FormType
	claims.SurveyURL = LauncherSchema.URL
}

// TokenError describes an error that can occur during JWT generation
type TokenError struct {
	// Err is a description of the error that occurred.
	Desc string

	// From is optionally the original error from which this one was caused.
	From error
}

func (e *TokenError) Error() string {
	if e == nil {
		return "<nil>"
	}
	err := e.Desc
	if e.From != nil {
		err += " (" + e.From.Error() + ")"
	}
	return err
}

// generateTokenFromClaims creates a token though encryption using the private and public keys
func generateTokenFromClaims(cl EqClaims) (string, *TokenError) {
	privateKeyResult, keyErr := loadSigningKey()
	if keyErr != nil {
		return "", &TokenError{Desc: "Error loading signing key", From: keyErr}
	}

	publicKeyResult, keyErr := loadEncryptionKey()
	if keyErr != nil {
		return "", &TokenError{Desc: "Error loading encryption key", From: keyErr}
	}

	opts := jose.SignerOptions{}
	opts.WithType("JWT")
	opts.WithHeader("kid", privateKeyResult.kid)

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: privateKeyResult.key}, &opts)
	if err != nil {
		return "", &TokenError{Desc: "Error creating JWT signer", From: err}
	}

	encryptor, err := jose.NewEncrypter(
		jose.A256GCM,
		jose.Recipient{Algorithm: jose.RSA_OAEP, Key: publicKeyResult.key, KeyID: publicKeyResult.kid},
		(&jose.EncrypterOptions{}).WithType("JWT").WithContentType("JWT"))

	if err != nil {
		return "", &TokenError{Desc: "Error creating JWT signer", From: err}
	}

	token, err := jwt.SignedAndEncrypted(signer, encryptor).Claims(cl).CompactSerialize()

	if err != nil {
		return "", &TokenError{Desc: "Error signing and encrypting JWT", From: err}
	}

	log.Println("Created signed/encrypted JWT:", token)

	return token, nil
}

// GenerateTokenFromDefaults coverts a set of DEFAULT values into a JWT
func GenerateTokenFromDefaults(url string, accountURL string) (token string, error string) {
	claims := EqClaims{}
	claims = generateDefaultClaims(accountURL)

	jwtClaims := GenerateJwtClaims()
	claims.Claims = jwtClaims

	launcherSchema, validationError := launcherSchemaFromURL(url)
	if validationError != "" {
		return "", validationError
	}
	addSchemaToClaims(&claims, launcherSchema)

	token, tokenError := generateTokenFromClaims(claims)
	if tokenError != nil {
		return token, fmt.Sprintf("GenerateTokenFromDefaults failed err: %v", tokenError)
	}

	return token, ""
}

// GenerateTokenFromPost coverts a set of POST values into a JWT
func GenerateTokenFromPost(postValues url.Values) (string, string) {
	log.Println("POST received: ", postValues)

	claims := EqClaims{}
	claims = generateClaimsFromPost(postValues)

	jwtClaims := GenerateJwtClaims()
	claims.Claims = jwtClaims

	schema := postValues.Get("schema")
	launcherSchema := surveys.FindSurveyByName(schema)
	addSchemaToClaims(&claims, launcherSchema)

	token, tokenError := generateTokenFromClaims(claims)
	if tokenError != nil {
		return token, fmt.Sprintf("GenerateTokenFromPost failed err: %v", tokenError)
	}

	return token, ""
}
