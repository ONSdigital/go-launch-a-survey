package authentication

import (
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"time"

	"github.com/satori/go.uuid"

	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"

	"github.com/ONSdigital/go-launch-a-survey/settings"
	"github.com/ONSdigital/go-launch-a-survey/surveys"
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

type eqClaims struct {
	jwt.Claims
	UserID                string       `json:"user_id"`
	EqID                  string       `json:"eq_id"`
	PeriodID              string       `json:"period_id"`
	PeriodStr             string       `json:"period_str"`
	CollectionExerciseSid string       `json:"collection_exercise_sid"`
	RuRef                 string       `json:"ru_ref"`
	RuName                string       `json:"ru_name"`
	RefPStartDate         string       `json:"ref_p_start_date"` // iso_8601_date
	RefPEndDate           string       `json:"ref_p_end_date,omitempty"` // iso_8601_date
	FormType              string       `json:"form_type"`
	SurveyURL              string       `json:"survey_url"`
	ReturnBy              string       `json:"return_by"`
	TradAs                string       `json:"trad_as,omitempty"`
	EmploymentDate        string       `json:"employment_date,omitempty"` // iso_8601_date
	RegionCode            string       `json:"region_code,omitempty"`
	LanguageCode          string       `json:"language_code,omitempty"`
	VariantFlags          variantFlags `json:"variant_flags,omitempty"`
	Roles                 string       `json:"roles,omitempty"`
	TxID                  string       `json:"tx_id,omitempty"`
}

type variantFlags struct {
	SexualIdentity string `json:"sexual_identity,omitempty"`
}

func generateClaims(postValues url.Values) (claims eqClaims) {
	issued := time.Now()
	expires := issued.Add(time.Minute * 10) // TODO: Support custom exp: r.PostForm.Get("exp")

	schema := postValues.Get("schema")
	launcherSchema := surveys.FindSurveyByName(schema)

	return eqClaims{
		Claims: jwt.Claims{
			IssuedAt: jwt.NewNumericDate(issued),
			Expiry:   jwt.NewNumericDate(expires),
			ID:       uuid.NewV4().String(),
		},
		EqID:                  launcherSchema.EqID,
		FormType:              launcherSchema.FormType,
		SurveyURL:			   launcherSchema.URL,
		UserID:                postValues.Get("user_id"),
		PeriodID:              postValues.Get("period_id"),
		PeriodStr:             postValues.Get("period_str"),
		CollectionExerciseSid: postValues.Get("collection_exercise_sid"),
		RuRef:          postValues.Get("ru_ref"),
		RuName:         postValues.Get("ru_name"),
		RefPStartDate:  postValues.Get("ref_p_start_date"),
		RefPEndDate:    postValues.Get("ref_p_end_date"),
		ReturnBy:       postValues.Get("return_by"),
		TradAs:         postValues.Get("trad_as"),
		EmploymentDate: postValues.Get("employment_date"),
		RegionCode:     postValues.Get("region_code"),
		LanguageCode:   postValues.Get("language_code"),
		TxID:           uuid.NewV4().String(),
		Roles:          postValues.Get("roles"),
		VariantFlags: variantFlags{
			SexualIdentity: postValues.Get("sexual_identity"),
		},
	}
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

// ConvertPostToToken coverts a set of POST values into a JWT
func ConvertPostToToken(postValues url.Values) (string, *TokenError) {
	log.Println("POST received...", postValues)

	cl := generateClaims(postValues)

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

	fmt.Printf("Created signed/encrypted JWT: %v", token)
	return token, nil
}
