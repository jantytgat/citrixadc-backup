FROM golang:1.17-alpine AS BUILD

WORKDIR $GOPATH/src/github.com/jantytgat/citrixadc-backup
COPY . .

RUN go get -d -v
RUN go install -v


RUN GOOS=linux GOARCH=amd64 go build -o /tmp/citrixadc-backup main.go


FROM gcr.io/distroless/base:latest

COPY --from=BUILD /tmp/citrixadc-backup /usr/local/bin/citrixadc-backup

CMD ["/usr/local/bin/citrixadc-backup"]