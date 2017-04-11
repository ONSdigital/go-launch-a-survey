FROM scratch

ENV GO_LAUNCH_A_SURVEY_LISTEN_HOST="0.0.0.0"
ENV GO_LAUNCH_A_SURVEY_LISTEN_PORT="8000"

EXPOSE 8000

COPY go-launch-a-survey /
COPY static/ /static/
COPY templates/ /templates/
COPY jwt-test-keys /jwt-test-keys/

ENTRYPOINT ["/go-launch-a-survey"]
