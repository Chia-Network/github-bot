FROM golang:1 as builder

COPY . /app
WORKDIR /app
RUN make build

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/bin/github-bot /github-bot

ENTRYPOINT ["/github-bot", "--config", "/config/config.yml"]
