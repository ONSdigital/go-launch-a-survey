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

### Docker
Built using https://github.com/CenturyLinkLabs/golang-builder to create a tiny Docker image.

To build and run, exposing the server on port 8000 locally:

```
docker run --rm -v "$(pwd):/src" -v /var/run/docker.sock:/var/run/docker.sock centurylink/golang-builder
docker run -p 8000:8000 go-launch-a-survey:latest
```

### Notes
* There are lots of TODOs in the code
* There are no unit tests yet
* Lots of tidying required.
* JWT spec based on http://ons-schema-definitions.readthedocs.io/en/latest/jwt_profile.html

### Settings

Environment Variable | Meaning | Default
---------------------|---------|--------
GO_LAUNCH_A_SURVEY_LISTEN_HOST|Host address  to listen on|0.0.0.0
GO_LAUNCH_A_SURVEY_LISTEN_PORT|Host port to listen on|8000
SURVEY_RUNNER_URL|URL of Survey Runner to re-direct to when launching a survey|http://localhost:5000
JWT_ENCRYPTION_KEY_PATH|Path to the JWT Encryption Key (PEM format)|jwt-test-keys/sdc-user-authentication-encryption-sr-public-key.pem
JWT_SIGNING_KEY_PATH|Path to the JWT Signing Key (PEM format)|jwt-test-keys/sdc-user-authentication-signing-rrm-private-key.pem


### Code Review
looks pretty good imo, the obvious things i'd mention you've already commented (e.g. using `interface{}` as return type, merging the two structs)

[4:30] 
probably want some `\n`s on `fmt.Printf`, since that doesn't include a newline by default (edited)

[4:31] 
move regex to top so it's only compiled once on startup, but doesn't really matter too much if performance isn't an issue

chriscollis [4:32 PM] 
thanks, good points

iankent [4:32 PM] 
i'd probably consider something like gorilla pat/mux and being more explicit with `GET`/`POST` - not too bad in a small app, but you'd be repeating the switch block a lot with something bigger

[4:32] 
also the `default` block should set the status code to `405`
pat and mux are both relatively minimal - pat is easier, but doesn't (imo) handle paths particularly well

nah, nobody needs unit tests :wink:

[4:35] 
could also vendor in deps (we use govendor, but there's plenty of vendoring tools) - there should be one as part of the go toolchain in the next few go releases

you could approach settings.go slightly differently, rather than using a sync.Once, you can either use a package level declaration, or an `init`

chriscollis [4:37 PM] 
hmm, ok, wasn’t sure about multi-thread issues on that stuff

[4:37] 
not that this app is threaded

[4:37] 
was part learning exercise

[4:37] 
(I have multi-threaded C++ background)

iankent [4:37 PM] 
`func init` is run when the package is first imported, so can only ever happen once on application startup


technically it is :smile: `http` package will spawn new goroutines for each new connection, and go by default will have one OS thread per core :slightly_smiling_face:

[4:39] 
i'd also just have a `GetSetting(name string)` func, and return the string (which is immutable) rather than a map (which isnt)

[4:40] 
but yeah, tbh i'm just being pedantic, it's actually very neat and idiomatic go code :+1: