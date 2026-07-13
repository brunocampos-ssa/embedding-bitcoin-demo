FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api
FROM alpine:3.22
RUN adduser -D app
USER app
WORKDIR /app
COPY --from=build /api /app/api
EXPOSE 8080
CMD ["/app/api"]
