FROM golang:1.24.12-bookworm AS builder
WORKDIR /src
COPY main.go .
RUN CGO_ENABLED=0 go build -o /oss-emulator main.go

FROM scratch
COPY --from=builder /oss-emulator /oss-emulator
EXPOSE 9000
ENTRYPOINT ["/oss-emulator"]
