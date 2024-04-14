FROM golang:1.21-alpine AS builder

WORKDIR /usr/local/src

RUN apk --no-cache add bash git make gcc gettext musl-dev

COPY ["go.mod", "go.sum", "./"]
RUN go mod download

COPY . ./
RUN go build -o ./cmd/ cmd/main.go

FROM alpine as runner

COPY --from=builder /usr/local/src/cmd/ /
COPY config.yaml /config.yaml
COPY internal/storage/storage.db /internal/storage/storage.db

CMD ["/main"]