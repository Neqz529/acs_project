FROM golang:1.22 AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o app

FROM gcr.io/distroless/base
WORKDIR /root/
COPY --from=builder /app/app .
CMD ["/root/app"]
