# LiSTgo

[Lietuviškai](README.lt.md)

Lithuanian Speech Transcription services

## About

Services for running the transcription process. It uses Mongo DB (for saving statuses) and RabbitMQ (for event bus).

---

## Build notes

1. Go to the base *listgo* dir

1. Get required packages:

    `go get ./...`

1. Install [librdkafka](https://github.com/confluentinc/confluent-kafka-go)
    `make install/librdkafka`

1. Generate mocks:

    `make generate/mocks`

---

## Testing source code

Run `make test/unit`

---

## Deploy note

For deploy options see [github.com/airenas/list](https://github.com/airenas/list)

---

## Author

Airenas Vaičiūnas

- [github.com/airenas](https://github.com/airenas)
- [linkedin.com/in/airenas](https://www.linkedin.com/in/airenas/)

---

## License

Copyright © 2022, [Airenas Vaičiūnas](https://github.com/airenas).
Released under the [The 3-Clause BSD License](LICENSE).

---
