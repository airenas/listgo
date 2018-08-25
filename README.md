# LiSTgo

Lithuanian Speech Transcription services

###About
Services for running the transcription process. It uses Mongo DB (for saving statuses) and RabbitMQ (for event bus)

---
###Build notes

1. Go to the base 'listgo' dir
2. Get required packages: 
    'go get ./...'
    'go get github.com/petergtz/pegomock/...'
    'go get github.com/smartystreets/goconvey/convey'
3. Generate mocks: 
    'go generate ./...'

---
###Testing source code
1. Go to the base 'listgo' dir
2. 'go test ./...'

---
###Deploy note
For deploy options see [bitbucket.org/airenas/list](https://bitbucket.org/airenas/list)

---
### Author

**Airenas Vaičiūnas**

* [bitbucket.org/airenas](https://bitbucket.org/airenas)

---
### License

Copyright © 2017, [Airenas Vaičiūnas](https://bitbucket.org/airenas).
Released under the [The 3-Clause BSD License](LICENSE).

---