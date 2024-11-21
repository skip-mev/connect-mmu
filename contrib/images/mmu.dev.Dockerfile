FROM golang:1.23-bullseye AS builder
WORKDIR /src/sauron
COPY . .

RUN go build -mod=vendor -o build/ ./...

FROM gcr.io/distroless/base-debian11:debug
COPY --from=builder /src/sauron/build/mmu /usr/local/bin/

# by default run indexing with the following entrypoint
# try to use a mounted indexing config

# to run the generate command, just change the entrypoint
ENTRYPOINT ["mmu"]
