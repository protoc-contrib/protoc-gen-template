# builder
FROM    golang:1.23-alpine AS builder
RUN     apk --no-cache add make git rsync libc-dev
WORKDIR /src
COPY    go.mod go.sum ./
RUN     go mod download
COPY    . .
RUN     CGO_ENABLED=0 go install -tags netgo -ldflags '-w -s' .

# runtime
FROM    znly/protoc:0.4.0
COPY    --from=builder  /go/bin/protoc-gen-go-template /go/bin/
ENV     PATH=$PATH:/go/bin
ENTRYPOINT []
