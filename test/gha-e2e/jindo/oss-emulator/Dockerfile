# golang:1.26.11-bookworm
FROM golang:1.26.11-bookworm@sha256:dc40a13b61304647ed7e5d581d246b86413dfc29d63c6d018310feb93cd98fc0 AS builder
WORKDIR /src
COPY main.go .
RUN CGO_ENABLED=0 go build -o /oss-emulator main.go

FROM scratch
COPY --from=builder /oss-emulator /oss-emulator
EXPOSE 9000
ENTRYPOINT ["/oss-emulator"]
