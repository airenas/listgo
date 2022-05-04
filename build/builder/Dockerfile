#########################################################################################
## docker will invoke this file from ./.. dir in order to access code
#########################################################################################
FROM golang:1.17 as builder

LABEL org.opencontainers.image.version=$BUILD_VERSION \
      org.opencontainers.image.authors="airenass@gmail.com" \
      name="list-builder" \
      maintainer="airenass@gmail.com" \
      vendor="airenass@gmail.com" \
      version=$BUILD_VERSION \
      release=$BUILD_VERSION \
      summary="This image is used to build listgo code" \
      description="This image is used to build listgo code" 
ARG BUILD_VERSION=0.1

WORKDIR /go/src/

RUN git clone --branch v1.1.0 https://github.com/edenhill/librdkafka.git && \
      cd librdkafka && ./configure --prefix /usr && make && make install && cd /go/src && rm -rf librdkafka

COPY go.mod /go/src/
COPY go.sum /go/src/
RUN go mod download
