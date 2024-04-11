FROM golang:1.22.2-alpine3.19

WORKDIR /app

COPY * /app

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o n2yo-proxy

ENV GIN_MODE release

EXPOSE 9443
ENTRYPOINT [ "./n2yo-proxy" ]
