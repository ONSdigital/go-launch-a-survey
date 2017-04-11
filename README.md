# Pre-Alpha: Go Launch a Survey!

### Building and Running
Install Go and ensure that your `GOPATH` env variable is set (usually it's `~/go`).

```
go get -d github.com/collisdigital/go-launch-a-survey/
cd $GOPATH/src/github.com/collisdigital/go-launch-a-survey/
go build
./go-launch-a-survey
```

Open http://localhost:8000/launch.html

### Notes
* There are lots of TODOs in the code
* There are no unit tests yet
* Lots of tidying required.
* JWT spec based on http://ons-schema-definitions.readthedocs.io/en/latest/jwt_profile.html

### Settings

Environment Variable | Meaning | Default
---------------------|---------|--------
GO_LAUNCH_A_SURVEY_LISTEN_HOST|Host address  to listen on|0.0.0.0
GO_LAUNCH_A_SURVEY_LISTEN_PORT|Host port to listen on|7999
SURVEY_RUNNER_URL|URL of Survey Runner to re-direct to when launching a survey|http://localhost:5000
JWT_ENCRYPTION_KEY_PATH|Path to the JWT Encryption Key (PEM format)|jwt-test-keys/sdc-user-authentication-encryption-sr-public-key.pem
JWT_SIGNING_KEY_PATH|Path to the JWT Signing Key (PEM format)|jwt-test-keys/sdc-user-authentication-signing-rrm-private-key.pem