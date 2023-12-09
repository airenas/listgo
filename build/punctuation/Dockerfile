#########################################################################################
## docker will invoke this file from ../.. dir in order to access code
#########################################################################################
FROM airenas/list-builder:0.3.0 as builder

ARG BUILD_VERSION=0.1

COPY . /go/src

RUN CGO_ENABLED=0 go build -o /go/bin/punctuationService -ldflags "-X main.version=$BUILD_VERSION" cmd/punctuationService/main.go
#####################################################################################
FROM alpine:3.15 as runner

ARG BUILD_VERSION=0.1

LABEL org.opencontainers.image.version=$BUILD_VERSION \
      org.opencontainers.image.authors="airenass@gmail.com" \
      name="list-punctuation" \
      maintainer="airenass@gmail.com" \
      vendor="airenass@gmail.com" \
      version=$BUILD_VERSION \
      release=$BUILD_VERSION \
      summary="This image is used to start the LiST punctuation service" \
      description="This image is used to start the LiST punctuation service" 

ENV UID=999
VOLUME /data
VOLUME /models
EXPOSE 8000

WORKDIR /app

RUN addgroup app && adduser -S -G app --uid $UID app
RUN mkdir -p /app && chown -R app:app /app   

COPY --from=builder /go/bin/punctuationService /app/
COPY build/punctuation/config.yaml /app/

RUN chown app:app /app/* /app
USER app

ENTRYPOINT ["./punctuationService"]
