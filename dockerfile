FROM golang

ADD . /go/src/github.com/e0m-ru/caldavreport

WORKDIR /go/src/github.com/e0m-ru/caldavreport

RUN  go mod tidy

RUN go install

ENTRYPOINT ["caldavreport"]

EXPOSE 8080