
# LiSTgo

Transkribatoriaus IT sprendimo servisų kodas

## Apie

Servisai valdantys transkribavimo procesą. Sistema naudoja Mongo DB (saugo būsenas) ir RabbitMQ (įvykių eilė).

## Prieš

Instaliuokite [go* (v >= 14.0)](https://golang.org/), [protoc](https://grpc.io/docs/protoc-installation/), [protoc-gen-go](https://grpc.io/docs/languages/go/quickstart/)

---

## Kompiliavimas

1. *Go* bibliotekų parsiuntimas:

    `go get ./...`

1. Instaliuokite [librdkafka](https://github.com/confluentinc/confluent-kafka-go)

    `git clone --branch v1.1.0 https://github.com/edenhill/librdkafka.git`

    `cd librdkafka`

    `./configure --prefix /usr`

    `make`

    `sudo make install`

1. Paruoškite tensorflow proto failus:

    `make generate/proto`

1. Sugeneruokite mock'us testavimui:

    `make generate/mocks`

---

## Testavimas

Vykdykite `go test ./...`

---

## Diegimo instrukcijos

Žr.: [github.com/airenas/list](https://github.com/airenas/list)

---

## Autorius

Airenas Vaičiūnas

- [github.com/airenas](https://github.com/airenas)
- [linkedin.com/in/airenas](https://www.linkedin.com/in/airenas/)

---

## Licencija

Copyright © 2020, [Airenas Vaičiūnas](https://github.com/airenas).

Released under the [The 3-Clause BSD License](LICENSE).

---
