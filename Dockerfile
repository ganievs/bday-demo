FROM golang:1.12-alpine as builder
ENV CGO_ENABLED 0
ENV GOOS linux
RUN apk add --no-cache git
WORKDIR /go/src/bday
ADD . /go/src/bday
RUN go get -t -v
RUN go test
RUN go build -a -installsuffix cgo -o bday .

FROM scratch
COPY --from=builder /go/src/bday/bday /bday
EXPOSE 8000
CMD ["/bday"]
