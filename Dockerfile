FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
ENV PORT=80
ENV ALLOWED_ORIGINS=https://super-tic-tac-toe-beta.vercel.app

COPY --from=builder /app/main .

EXPOSE 80
CMD ["./main"]