FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

# Install our build tools
RUN apk add --update ca-certificates

WORKDIR /go/src/app

COPY . ./

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o bin/node-undertaker github.com/dbschenker/node-undertaker/cmd/node-undertaker

RUN echo "nonroot:x:1337:1337:nonroot:/nonroot:/usr/sbin/nologin" > /etc_passwd

FROM --platform=$BUILDPLATFORM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/app/bin/* /
COPY --from=builder /etc_passwd /etc/passwd

USER nonroot

ENTRYPOINT ["/node-undertaker"]
