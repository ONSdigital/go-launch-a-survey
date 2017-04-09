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