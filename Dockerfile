FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache poppler-utils ca-certificates
COPY --from=build /out/server /app/server
ENV ADDR=:8090
EXPOSE 8090
ENTRYPOINT ["/app/server"]
