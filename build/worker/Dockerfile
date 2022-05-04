#########################################################################################
## docker will invoke this file from ../.. dir in order to access code
#########################################################################################
FROM airenas/list-builder:0.2.388-0c1d460 as builder

ARG BUILD_VERSION=0.1

COPY . /go/src

RUN CGO_ENABLED=0 go build -o /go/bin/cmdWorkerService -ldflags "-X main.version=$BUILD_VERSION" cmd/cmdWorkerService/main.go
#####################################################################################
FROM alpine:3.15 as runner

RUN apk add --no-cache make

ARG BUILD_VERSION=0.1

LABEL org.opencontainers.image.version=$BUILD_VERSION \
      org.opencontainers.image.authors="airenass@gmail.com" \
      name="list-manager" \
      maintainer="airenass@gmail.com" \
      vendor="airenass@gmail.com" \
      version=$BUILD_VERSION \
      release=$BUILD_VERSION \
      summary="This image is used to start the LiST worker" \
      description="This image is used to start the LiST worker" 

ENV UID=999
WORKDIR /app

RUN addgroup app && adduser -S -G app --uid $UID app
RUN mkdir -p /app && chown -R app:app /app   

COPY --from=builder /go/bin/cmdWorkerService /app/

RUN chown app:app /app/* /app
USER app

ENTRYPOINT ["./cmdWorkerService"]
