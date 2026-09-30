FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o marketplace-api ./cmd/main.go

FROM gcr.io/distroless/static-debian13

WORKDIR /app

COPY --from=builder /app/marketplace-api .
EXPOSE 8080

CMD ["./marketplace-api"]
