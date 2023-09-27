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

RUN go build -v -o server ./cmd/fireactions

RUN adduser --disabled-password --gecos '' appuser && \
    chown -R appuser /app

FROM scratch

COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /app/server /app/server

USER appuser

ENTRYPOINT ["/app/fireactions", "server"]
