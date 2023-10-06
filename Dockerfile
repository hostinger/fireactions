FROM golang:1.20 as build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

ENV GO111MODULE=on \
    CGO_ENABLED=0  \
    GOOS=linux     \
    GOARCH=amd64

RUN go build -v -o fireactions .

RUN adduser --disabled-password --uid 1000 --gecos '' appuser && \
    chown -R appuser /app

FROM alpine:3.18.4

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /app/fireactions /usr/bin/fireactions

USER appuser

ENTRYPOINT ["/usr/bin/fireactions", "server"]
