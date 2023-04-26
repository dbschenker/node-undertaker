FROM --platform=$BUILDPLATFORM golang:1.20-alpine as builder

ARG TARGETOS
ARG TARGETARCH

# Install our build tools
RUN apk add --update ca-certificates

WORKDIR /go/src/app

COPY . ./

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o bin/node-undertaker gilds-git.signintra.com/aws-dctf/kubernetes/node-undertaker/cmd/node-undertaker

FROM --platform=$BUILDPLATFORM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/app/bin/* /

ENTRYPOINT ["/node-undertaker"]
