# build image
FROM golang:1.14-alpine AS build-env

# install build tools
RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

# build
WORKDIR /app
COPY . .
WORKDIR /app/main
RUN go build -mod=vendor -o main .

# default env
ENV PORT 18080

# distribution image
FROM alpine:3.11

# add CAs
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=build-env /app/app/main /app/app

# start
CMD ["./app", "start"]