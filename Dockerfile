FROM golang:1.6

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

# this will ideally be built by the ONBUILD below ;)
EXPOSE 3000
CMD ["go-wrapper", "run"]

COPY . /go/src/app
RUN go-wrapper download
RUN go-wrapper install
