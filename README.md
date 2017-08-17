# Alpha: Go Launch a Survey!

### Building and Running
Install Go and ensure that your `GOPATH` env variable is set (usually it's `~/go`).

Note this app uses govendor (https://github.com/kardianos/govendor) to manage its dependencies.

```
go get -d github.com/ONSdigital/go-launch-a-survey/
cd $GOPATH/src/github.com/ONSdigital/go-launch-a-survey/
go build
./go-launch-a-survey
```

Open http://localhost:8000/

### Docker
Built using https://github.com/CenturyLinkLabs/golang-builder to create a tiny Docker image.

To build and run, exposing the server on port 8000 locally:

```
docker run --rm -v "$(pwd):/src" -v /var/run/docker.sock:/var/run/docker.sock centurylink/golang-builder
docker run -it -p 8000:8000 go-launch-a-survey:latest
```

You can also run a Survey Register for launcher to load Schemas from 
```
docker run -it -p 8080:8080 onsdigital/eq-survey-register:simple-rest-api
```

### Notes
* There are no unit tests yet
* JWT spec based on http://ons-schema-definitions.readthedocs.io/en/latest/jwt_profile.html

### Settings
Environment Variable | Meaning | Default
---------------------|---------|--------
GO_LAUNCH_A_SURVEY_LISTEN_HOST|Host address  to listen on|0.0.0.0
GO_LAUNCH_A_SURVEY_LISTEN_PORT|Host port to listen on|8000
SURVEY_RUNNER_URL|URL of Survey Runner to re-direct to when launching a survey|http://localhost:5000
SURVEY_REGISTER_URL|URL of eq-survey-register to load schema list from |http://localhost:8080
JWT_ENCRYPTION_KEY_PATH|Path to the JWT Encryption Key (PEM format)|jwt-test-keys/sdc-user-authentication-encryption-sr-public-key.pem
JWT_SIGNING_KEY_PATH|Path to the JWT Signing Key (PEM format)|jwt-test-keys/sdc-user-authentication-signing-rrm-private-key.pem
