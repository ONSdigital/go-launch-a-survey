FROM alpine

RUN apk --no-cache update && \
    apk --no-cache add python py-pip py-setuptools ca-certificates groff less && \
    pip --no-cache-dir install awscli && \
    rm -rf /var/cache/apk/*

ENV GO_LAUNCH_A_SURVEY_LISTEN_HOST="0.0.0.0"
ENV GO_LAUNCH_A_SURVEY_LISTEN_PORT="8000"

EXPOSE 8000

COPY docker-entrypoint.sh /
COPY static/ /static/
COPY templates/ /templates/
COPY jwt-test-keys /jwt-test-keys/
COPY go-launch-a-survey /

ENTRYPOINT ["sh", "/docker-entrypoint.sh"]
