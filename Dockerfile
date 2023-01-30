FROM golang:latest AS builder
LABEL stage=gobuilder
ENV CGO_ENABLED 0
ENV GOOS linux
WORKDIR /build
COPY ./ /build
RUN go mod download
RUN go build -ldflags="-s -w" -o b2b-chat /build/cmd/b2b-chat/main.go

FROM alpine
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /build/b2b-chat .
COPY --from=builder /build/config.yaml .
CMD ["./b2b-chat"]