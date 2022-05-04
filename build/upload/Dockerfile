#########################################################################################
## docker will invoke this file from ../.. dir in order to access code
#########################################################################################
FROM airenas/list-builder:0.1.382-cdd1caf as builder


COPY . /go/src

ARG BUILD_VERSION=0.1

# copy src
COPY . /go/src

RUN CGO_ENABLED=0 go build -o /go/bin/uploadService -ldflags "-X main.version=$BUILD_VERSION" cmd/uploadService/main.go
#####################################################################################
FROM alpine:3.15 as runner

ARG BUILD_VERSION=0.1

LABEL org.opencontainers.image.version=$BUILD_VERSION \
      org.opencontainers.image.authors="airenass@gmail.com" \
      name="list-upload" \
      maintainer="airenass@gmail.com" \
      vendor="airenass@gmail.com" \
      version=$BUILD_VERSION \
      release=$BUILD_VERSION \
      summary="This image is used to start the LiST upload service" \
      description="This image is used to start the LiST upload service" 

ENV UID=999
VOLUME /data
VOLUME /models
EXPOSE 8000

WORKDIR /app

RUN addgroup app && adduser -S -G app --uid $UID app
RUN mkdir -p /app && chown -R app:app /app   

COPY --from=builder /go/bin/uploadService /app/
COPY build/upload/config.yaml /app/

RUN chown app:app /app/* /app
USER app

ENTRYPOINT ["./uploadService"]
