FROM golang:1.23-bullseye AS builder
USER root
WORKDIR /src/connect-mmu
RUN apt-get update && apt-get install -y curl && apt-get install jq -y
COPY . .

RUN go build -o build/ ./...
RUN make install-sentry

# install slinky v1.0.12 for the OS + architecture of the current machine
COPY scripts/install_slinky.sh /tmp/
RUN chmod +x /tmp/install_slinky.sh && /tmp/install_slinky.sh $(uname -s | tr '[:upper:]' '[:lower:]') $(uname -m) 

# install a bunch of connect versions for the OS + architecture of the current machine.
# the script will install v2.0.0 and onwards.
COPY scripts/install_all_connects.sh /tmp/
RUN chmod +x /tmp/install_all_connects.sh && /tmp/install_all_connects.sh $(uname -s | tr '[:upper:]' '[:lower:]') $(uname -m) 

FROM ubuntu:rolling
COPY --from=builder /src/connect-mmu/build/mmu /usr/local/bin/
COPY --from=builder /usr/local/bin/slinky /usr/local/bin/
COPY --from=builder /go/bin/sentry /usr/local/bin/
# Copy all connect binaries
COPY --from=builder /usr/local/bin/connect-* /usr/local/bin/
COPY --from=builder /usr/local/bin/connect /usr/local/bin/
# symlink slinky -> connect-1.0.12
RUN ln -s /usr/local/bin/slinky /usr/local/bin/connect-1.0.12

EXPOSE 8002

WORKDIR /usr/local/bin/

RUN apt-get update && apt-get install ca-certificates -y

ENTRYPOINT ["mmu"]