FROM golang:1.22-alpine AS build

WORKDIR /src/backend
COPY backend/go.mod ./
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/gheymatchi-api ./cmd/api

FROM alpine:3.20

RUN adduser -D -H -u 10001 appuser
COPY --from=build /out/gheymatchi-api /usr/local/bin/gheymatchi-api

ENV APP_ENV=local
ENV HTTP_ADDR=:8080

EXPOSE 8080
USER appuser

ENTRYPOINT ["/usr/local/bin/gheymatchi-api"]
