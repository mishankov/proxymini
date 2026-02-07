FROM oven/bun:1 AS webui-builder
WORKDIR /app
COPY webui ./webui
COPY webapp ./webapp
WORKDIR /app/webui
RUN bun install --frozen-lockfile
RUN bun run build

FROM golang:1.24 AS builder
WORKDIR /app
COPY . .
COPY --from=webui-builder /app/webapp/static /app/webapp/static
RUN CGO_ENABLED=0 go build -o /app/proxymini ./cmd/proxymini

FROM alpine:3
COPY --from=builder /app/proxymini /app/proxymini
RUN mkdir /app/data
ENV PROXYMINI_DB="/app/data/rl.db"
WORKDIR /app
CMD ["./proxymini"]
