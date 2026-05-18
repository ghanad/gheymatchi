FROM golang:1.22-alpine AS build

WORKDIR /src/backend
COPY backend/go.mod ./
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/gheymatchi-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/gheymatchi-migrate ./cmd/migrate
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/gheymatchi-worker ./cmd/worker

FROM alpine:3.20

RUN adduser -D -H -u 10001 appuser && mkdir -p /data && chown appuser /data
WORKDIR /app
COPY --from=build /out/gheymatchi-api /usr/local/bin/gheymatchi-api
COPY --from=build /out/gheymatchi-migrate /usr/local/bin/gheymatchi-migrate
COPY --from=build /out/gheymatchi-worker /usr/local/bin/gheymatchi-worker
COPY backend/migrations ./migrations

ENV APP_ENV=local
ENV HTTP_ADDR=:8080
ENV DB_PATH=/data/gheymatchi.db

EXPOSE 8080
USER appuser

CMD ["/usr/local/bin/gheymatchi-api"]
