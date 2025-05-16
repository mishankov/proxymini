FROM golang:1.24 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o /app/proxymini ./cmd/proxymini

FROM alpine:3
COPY --from=builder /app/proxymini /app/proxymini
RUN mkdir /app/data
ENV PROXYMINI_DB="/app/data/rl.db"
WORKDIR /app
CMD ["./proxymini"]
