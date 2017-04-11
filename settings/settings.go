package settings

import (
	"os"
	"sync"
)

type Settings map[string]string

var settings Settings
var once sync.Once

func getSetting(settings Settings, key string, defaultValue string) {
	if value, present := os.LookupEnv(key); present {
		settings[key] = value
	} else {
		settings[key] = defaultValue
	}
}

func GetSettings() Settings {
	once.Do(func() {
		settings = make(Settings)
		getSetting(settings, "GO_LAUNCH_A_SURVEY_LISTEN_HOST", "0.0.0.0")
		getSetting(settings, "GO_LAUNCH_A_SURVEY_LISTEN_PORT", "8000")
		getSetting(settings, "SURVEY_RUNNER_URL", "http://localhost:5000")
		getSetting(settings, "JWT_ENCRYPTION_KEY_PATH", "jwt-test-keys/sdc-user-authentication-encryption-sr-public-key.pem")
		getSetting(settings, "JWT_SIGNING_KEY_PATH", "jwt-test-keys/sdc-user-authentication-signing-rrm-private-key.pem")
	})
	return settings
}
