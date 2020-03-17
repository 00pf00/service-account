FROM golang:1.13
ADD . /go/src/00pf00/service-account
WORKDIR /go/src/00pf00/service-account
COPY . $WORKDIR
RUN make