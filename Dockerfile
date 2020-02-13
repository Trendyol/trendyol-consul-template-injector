FROM golang:1.13-buster as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./*.go

FROM gcr.io/distroless/base-debian10
COPY --from=build /app/main ./
CMD ["/main"]