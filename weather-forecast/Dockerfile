FROM golang:latest as build-env
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ./app

FROM alpine:latest
ENV API_URL=""
ENV REDIS_CONNECTION_STRING=""
WORKDIR /app
COPY --from=build-env /build/app .
RUN ls 
CMD ["./app"]