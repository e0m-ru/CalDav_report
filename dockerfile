FROM golang

ADD . /go/src/github.com/e0m-ru/echoserver

WORKDIR /go/src/github.com/e0m-ru/echoserver

RUN  go mod tidy

RUN go install

ENTRYPOINT ["echoserver"]

EXPOSE 8080