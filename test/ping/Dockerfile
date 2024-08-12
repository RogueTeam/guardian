FROM golang:alpine AS build

WORKDIR /source
COPY go.mod .
RUN go mod download
COPY . .
RUN go build -o /build/ping ./test/ping

FROM alpine:latest

COPY --from=build /build/ping /ping

CMD /ping