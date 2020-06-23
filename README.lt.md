
# LiSTgo

Transkribatoriaus IT sprendimo servisų kodas

## Apie

Servisai valdantys transkribavimo procesą. Sistema naudoja Mongo DB (saugo būsenas) ir RabbitMQ (įvykių eilė).

## Prieš

Instaliuokite *go* (v >= 14.0), *protoc*, [protoc-gen-go](https://grpc.io/docs/languages/go/quickstart/)

---

## Kompiliavimas

1. *Go* bibliotekų parsiuntimas:

    `go get ./...`

    `go get github.com/petergtz/pegomock/...`

1. Instaliuokite [librdkafka](https://github.com/confluentinc/confluent-kafka-go)

    `git clone --branch v1.1.0 https://github.com/edenhill/librdkafka.git`

    `cd librdkafka`

    `./configure --prefix /usr`

    `make`

    `sudo make install`

1. Paruoškite tensorflow proto filus:

    `cd build && ./prepareTFProto.sh`

1. Sugeneruokite mock'us testavimui:

    `go generate ./...`

---

## Testavimas

Vykdykite `go test ./...`

---

## Diegimo instrukcijos

Žr.: [bitbucket.org/airenas/list](https://bitbucket.org/airenas/list)

---

## Autorius

Airenas Vaičiūnas

- [bitbucket.org/airenas](https://bitbucket.org/airenas)
- [linkedin.com/in/airenas](https://www.linkedin.com/in/airenas/)

---

## Licencija

Copyright © 2020, [Airenas Vaičiūnas](https://bitbucket.org/airenas).

Released under the [The 3-Clause BSD License](LICENSE).

---
