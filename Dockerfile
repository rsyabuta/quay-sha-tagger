FROM golang:1.6-alpine

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

# this will ideally be built by the ONBUILD below ;)
EXPOSE 3000
CMD ["quay-sha-tagger"]

COPY . /go/src/app
RUN go build -v
