# Source: https://github.com/rebuy-de/golang-template
# Version: 1.3.1

FROM golang:1.9-alpine

RUN apk add --no-cache git make

# Configure Go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin

# Install Go Tools
RUN go get -u github.com/golang/lint/golint
RUN go get -u github.com/golang/dep/cmd/dep

COPY . /go/src/github.com/rebuy-de/exporter-merger
WORKDIR /go/src/github.com/rebuy-de/exporter-merger
RUN CGO_ENABLED=0 make install

ENTRYPOINT ["/go/bin/exporter-merger"]
