FROM oven/bun:1 AS webui-builder
COPY webui /app/webui
WORKDIR /app/webui
RUN bun install --frozen-lockfile
RUN bun run build

FROM golang:1.25 AS backend-builder
WORKDIR /app
COPY . .
COPY --from=webui-builder /app/webui/build /app/webapp
RUN CGO_ENABLED=0 go build -o /app/proxymini ./cmd/proxymini

FROM alpine:3
COPY --from=backend-builder /app/proxymini /app/proxymini
RUN mkdir /app/data
ENV PROXYMINI_DB="/app/data/rl.db"
WORKDIR /app
CMD ["./proxymini", "run"]
