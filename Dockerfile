# ── 1) Monta a interface web (React) ──────────────────────────
FROM node:22-alpine AS client
WORKDIR /app/client
COPY client/package.json client/package-lock.json ./
RUN npm ci
COPY client/ ./
RUN npm run build

# ── 2) Compila o servidor (Go) ────────────────────────────────
FROM golang:1.26-alpine AS server
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/wacalls ./cmd/server

# ── 3) Imagem final ───────────────────────────────────────────
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 wacalls
WORKDIR /app
COPY --from=server /out/wacalls /app/wacalls
COPY --from=client /app/client/dist /app/client/dist
RUN mkdir -p /data && chown -R wacalls:wacalls /data /app
USER wacalls
ENV TZ=America/Sao_Paulo
EXPOSE 8080

# A sessao do WhatsApp fica em /data — precisa de volume, senao some no redeploy.
CMD ["/app/wacalls", "-addr", ":8080", "-db", "/data/wacalls.db", "-static", "/app/client/dist", "-debug"]
