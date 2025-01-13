# Build
FROM golang:alpine AS build

RUN apk add --update git make build-base && \
    rm -rf /var/cache/apk/*

WORKDIR /src/
COPY . /src/
RUN go build -o gomarks

# Runtime
FROM alpine:latest

RUN mkdir /app
RUN mkdir /data
RUN mkdir /static
COPY --from=build /src/gomarks /app
COPY --from=build /src/static /static

VOLUME /data

EXPOSE 8080/tcp

ENTRYPOINT ["/app/gomarks"]
CMD []
