package authentication

import (
	"fmt"
	"log"
	"time"
	"errors"
	"regexp"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"net/url"

	"github.com/satori/go.uuid"

	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"

	"github.com/collisdigital/go-launch-a-survey/settings"
)

func loadEncryptionKey() (key *rsa.PublicKey, error error) {

	encryptionKeyPath := settings.GetSetting("JWT_ENCRYPTION_KEY_PATH")
	if keyData, err := ioutil.ReadFile(encryptionKeyPath); err == nil {
		block, _ := pem.Decode(keyData)
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err == nil {
			return pub.(*rsa.PublicKey), nil
		}
	}
	return nil, errors.New("Failed to load encryption key")
}

// TODO: Add support for password protected private keys
func loadSigningKey() (key *rsa.PrivateKey, error error) {
	signingKeyPath := settings.GetSetting("JWT_SIGNING_KEY_PATH")
	if keyData, err := ioutil.ReadFile(signingKeyPath); err == nil {
		block, _ := pem.Decode(keyData)
		priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err == nil {
			return priv, nil
		}
	}
	return nil, errors.New("Failed to load signing key")
}

type eqClaims struct {
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

var eqIdFormTypeRegex = regexp.MustCompile(`^(?P<eq_id>[a-z0-9]+)_(?P<form_type>\w+)\.json`)

func extractEqIdFormType(schema string) (eqId, formType string) {
	match := eqIdFormTypeRegex.FindStringSubmatch(schema)
	if match != nil {
		eqId = match[1]
		formType = match[2]
	}
	return
}

func generateClaims(postValues url.Values) (claims eqClaims) {
	issued := time.Now()
	expires := issued.Add(time.Minute * 10) // TODO: Support custom exp: r.PostForm.Get("exp")

	schema := postValues.Get("schema")
	eqId, formType := extractEqIdFormType(schema)

	return eqClaims{
		Claims: jwt.Claims{
			IssuedAt: jwt.NewNumericDate(issued),
			Expiry:   jwt.NewNumericDate(expires),
			ID:       uuid.NewV4().String(),
		},
		EqId:                  eqId,
		FormType:              formType,
		UserId:                postValues.Get("user_id"),
		PeriodId:              postValues.Get("period_id"),
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
		TxId:           uuid.NewV4().String(),
		// TODO: Support: VariantFlags
		// TODO: Support: Roles
	}
}

func ConvertPostToToken(postValues url.Values) (token string, err error) {
	log.Println("POST received...", postValues)

	cl := generateClaims(postValues)

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
