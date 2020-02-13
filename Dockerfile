FROM golang:1.14rc1-alpine3.11 as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main ./*.go

FROM alpine
COPY --from=build /app/main .
CMD ["./main"]