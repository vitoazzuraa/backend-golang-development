FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY api-student ./api-student

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/app ./api-student

FROM alpine:3.20

RUN adduser -D -u 10001 app

WORKDIR /app

COPY --from=build /out/app ./app
COPY --from=build /src/api-student/migrations ./migrations

USER app

ENTRYPOINT ["./app"]
