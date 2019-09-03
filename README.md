# LiSTgo

Lithuanian Speech Transcription services

###About

Services for running the transcription process. It uses Mongo DB (for saving statuses) and RabbitMQ (for event bus).

---
###Build notes

1. Go to the base *listgo* dir
2. Get required packages: 
  -  `go get ./...`
  -  `go get github.com/petergtz/pegomock/...`
3. Install [librdkafka](https://github.com/confluentinc/confluent-kafka-go)
```bash
  git clone --branch v1.1.0 https://github.com/edenhill/librdkafka.git
  cd librdkafka
  ./configure --prefix /usr
  make
  sudo make install
```
4. Prepare tensorflow proto files for go
```bash
  cd build
  ./prepareTFProto.sh
```
5. Generate mocks: 
    `go generate ./...`

---
###Testing source code

1. Go to the base *listgo* dir
2. `go test ./...`

---
###Deploy note

For deploy options see [bitbucket.org/airenas/list](https://bitbucket.org/airenas/list)

---
### Author

**Airenas Vaičiūnas**

* [bitbucket.org/airenas](https://bitbucket.org/airenas)
* [linkedin.com/in/airenas](https://www.linkedin.com/in/airenas/)


---
### License

Copyright © 2019, [Airenas Vaičiūnas](https://bitbucket.org/airenas).
Released under the [The 3-Clause BSD License](LICENSE).

---