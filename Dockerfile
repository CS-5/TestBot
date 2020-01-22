# Build Go App
FROM golang:alpine
RUN apk add --no-cache git gcc g++
ENV CGO_ENABLED=1
ENV GOOS=linux
WORKDIR /app
COPY *.go go.mod go.sum ./
RUN go build -o zeroxsix .

# Build Docker Image
FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=0 /app/zeroxsix .

CMD ./zeroxsix