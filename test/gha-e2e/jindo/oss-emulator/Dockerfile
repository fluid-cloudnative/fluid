# golang:1.26.10-bookworm
FROM golang:1.26.10-bookworm@sha256:23ce71337b726608afb1bc00986e630a05034a87a87ea1fc2cd1ae5cf33387eb AS builder
WORKDIR /src
COPY main.go .
RUN CGO_ENABLED=0 go build -o /oss-emulator main.go

FROM scratch
COPY --from=builder /oss-emulator /oss-emulator
EXPOSE 9000
ENTRYPOINT ["/oss-emulator"]
