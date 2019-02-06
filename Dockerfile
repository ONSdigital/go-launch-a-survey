# Start from golang v1.8 base image
FROM golang:1.8 as builder

WORKDIR /go/src/github.com/ONSdigital/go-launch-a-survey

COPY . .

# Download dependencies
RUN go get -u github.com/golang/dep/cmd/dep
RUN $GOPATH/bin/dep ensure

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/go-launch-a-survey .

######## Start a new stage from scratch #######
FROM alpine:latest  

RUN apk --no-cache add ca-certificates

# Copy the Pre-built binary file and entry point from the previous stage
COPY --from=builder /go/bin/go-launch-a-survey .
COPY docker-entrypoint.sh .
COPY static/ /static/
COPY templates/ /templates/
COPY jwt-test-keys /jwt-test-keys/

EXPOSE 8000

ENTRYPOINT ["sh", "/docker-entrypoint.sh"]
