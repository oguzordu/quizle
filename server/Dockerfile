FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/quizle ./cmd/quizle

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/quizle ./quizle
COPY --from=build /src/data ./data
EXPOSE 8080
ENTRYPOINT ["./quizle"]
