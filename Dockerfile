# build image
FROM golang:1.14 AS build-env

# build
WORKDIR /app
COPY . .
WORKDIR /app/main

# run test & build, ensure binary is unit test passed
RUN go test -v -mod=vendor ./...
RUN go build -mod=vendor -o main .

# distribution image
FROM alpine:3.11

# add CAs
RUN apk --no-cache add ca-certificates bash

WORKDIR /app
COPY --from=build-env /app/main/main /app/app

# default env
ENV ADDR "0.0.0.0:18080"

# default EXPOSE 18080
EXPOSE 18080

# start
CMD ["./app", "start"]