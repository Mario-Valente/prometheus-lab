FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY go.sum* ./  # Conditional copy - matches go.sum if it exists, does not fail if missing
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o prometheus-app .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/prometheus-app .
EXPOSE 8080
CMD ["./prometheus-app"]
