package settings

import (
	"os"
	"net"
	"fmt"
	"strings"
)


var _settings map[string]string

func setSetting(key string, defaultValue string) {
	if value, present := os.LookupEnv(key); present {
		_settings[key] = value
	} else {
		_settings[key] = defaultValue
	}
}

func init() {
	addrs, err := net.InterfaceAddrs()
	
	if err != nil {
		fmt.Println(err)
	}

	var currentIP string

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				currentIP = ipnet.IP.String()
			}
		}
	}


	_settings = make(map[string]string)
	setSetting("GO_LAUNCH_A_SURVEY_LISTEN_HOST", "0.0.0.0")
	setSetting("GO_LAUNCH_A_SURVEY_LISTEN_PORT", "8000")
	setSetting("SURVEY_RUNNER_URL", strings.Replace("http://localhost:5000", "localhost", currentIP, -1))
	setSetting("SURVEY_RUNNER_SCHEMA_URL", Get("SURVEY_RUNNER_URL"))
	setSetting("SCHEMA_VALIDATOR_URL", "")
	setSetting("SURVEY_REGISTER_URL", "")
	setSetting("JWT_ENCRYPTION_KEY_PATH", "jwt-test-keys/sdc-user-authentication-encryption-sr-public-key.pem")
	setSetting("JWT_SIGNING_KEY_PATH", "jwt-test-keys/sdc-user-authentication-signing-rrm-private-key.pem")
}

// Get returns the value for the specified named setting
func Get(name string) string {
	return _settings[name]
}
