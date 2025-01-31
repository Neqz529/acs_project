FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o app

FROM debian:stable-slim AS final
WORKDIR /root/

RUN apt-get update && apt-get install -y ffmpeg && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/app .
CMD ["/root/app"]
