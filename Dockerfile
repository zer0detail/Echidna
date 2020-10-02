FROM golang:latest AS build-env
WORKDIR /src
ENV GO111MODULE=on
COPY go.mod /src/
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o Echidna -ldflags="-s -w" -gcflags="all=-trimpath=/src" -asmflags="all=-trimpath=/src"

FROM alpine:latest

RUN mkdir -p /app \
    && adduser -D Echidna \
    && chown -R Echidna:Echidna /app

USER Echidna
WORKDIR /app

ARG APP_VOLUME=/app
VOLUME [${APP_VOLUME}]

COPY --from=build-env /src/Echidna .

ENTRYPOINT [ "./Echidna" ]