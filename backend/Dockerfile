# ---------- Stage 1: build a static Go binary ----------
FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY go.mod ./
COPY *.go ./

# CGO_ENABLED=0 produces a fully static binary that runs on scratch/alpine
# with no dynamic library dependencies.
RUN CGO_ENABLED=0 GOOS=linux go build -o cloudtask-backend .

# ---------- Stage 2: minimal runtime image ----------
FROM alpine:3.19 AS runner
WORKDIR /app

RUN apk add --no-cache ca-certificates wget && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/cloudtask-backend .

ENV PORT=8080
USER appuser
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./cloudtask-backend"]
