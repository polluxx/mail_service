FROM golang:latest

#RUN mkdir /go/app

ADD src /go/src
ADD main.go /go/

WORKDIR /go/

RUN go get github.com/gin-gonic/gin && go get github.com/ahmetalpbalkan/go-linq && go get github.com/mhale/smtpd

RUN go build .

EXPOSE 8080 2525

ENTRYPOINT /go/go