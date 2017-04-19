package settings

import "os"

var _settings map[string]string

func setSetting(key string, defaultValue string) {
	if value, present := os.LookupEnv(key); present {
		_settings[key] = value
	} else {
		_settings[key] = defaultValue
	}
}

func init() {
	_settings = make(map[string]string)
	setSetting("GO_LAUNCH_A_SURVEY_LISTEN_HOST", "0.0.0.0")
	setSetting("GO_LAUNCH_A_SURVEY_LISTEN_PORT", "8000")
	setSetting("SURVEY_RUNNER_URL", "http://localhost:5000")
	setSetting("JWT_ENCRYPTION_KEY_PATH", "jwt-test-keys/sdc-user-authentication-encryption-sr-public-key.pem")
	setSetting("JWT_SIGNING_KEY_PATH", "jwt-test-keys/sdc-user-authentication-signing-rrm-private-key.pem")
}

func GetSetting(name string) string {
	return _settings[name]
}