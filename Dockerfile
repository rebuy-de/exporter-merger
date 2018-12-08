# Source: https://github.com/rebuy-de/golang-template
# Version: 1.3.1

FROM golang:1.9-alpine

RUN apk add --no-cache git make

# Configure Go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin

# Install Go Tools
RUN go get -u golang.org/x/lint/golint
RUN go get -u github.com/golang/dep/cmd/dep

COPY Dockerfile Gopkg.lock Gopkg.toml LICENSE Makefile README.md golang.mk main.go merger.yaml /go/src/github.com/rebuy-de/exporter-merger/
COPY .git /go/src/github.com/rebuy-de/exporter-merger/.git/
COPY cmd /go/src/github.com/rebuy-de/exporter-merger/cmd/
WORKDIR /go/src/github.com/rebuy-de/exporter-merger
RUN CGO_ENABLED=0 make install

COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
