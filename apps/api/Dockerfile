FROM golang:1.25-alpine AS build
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN go mod download
COPY apps/api ./apps/api
COPY database ./database
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./apps/api/cmd/server \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./apps/api/cmd/migrate

FROM alpine:3.24
RUN apk add --no-cache ca-certificates && addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /out/api /out/migrate /usr/local/bin/
COPY database/migrations ./database/migrations
COPY docs/openapi.yaml ./docs/openapi.yaml
USER app
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api"]
