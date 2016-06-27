FROM golang:1.6-alpine

RUN apk add --no-cache git && \
    mkdir -p /go/src/app
WORKDIR /go/src/app

# this will ideally be built by the ONBUILD below ;)
EXPOSE 3000
CMD ["go-wrapper", "run"]

COPY . /go/src/app
RUN go get -v -u ./...
RUN go build -v
