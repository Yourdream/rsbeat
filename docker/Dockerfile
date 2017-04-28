FROM alpine:3.4

RUN apk add --update curl && \
    rm -rf /var/cache/apk/*

RUN curl -fsSL 'https://github.com/Yourdream/rsbeat/releases/download/v5.3.2/rsbeat-linux-amd64' -o rsbeat

RUN pwd && mkdir -p /app && mv rsbeat /app

ENTRYPOINT ["/app/rsbeat","-strict.perms=false"]

WORKDIR /app
