## AllSpark

AllSpark is a simple building block for quickly building web microservice deployments for demo purposes.

It supports at most one workload and arbitrary number of subsequent requests to other services that will be called on a single request.

### Workload

The used workload can be set using the `WORKLOAD` environment variable.

Currently the following workloads are supported:

#### PI

Calculates PI to generate CPU and memory usage

Available options:

- PI_COUNT - how many goroutines to use

#### ECHO

Simply echos back a set string

Available options:

- EHCO_STR - string to echo back

#### AIRPORTS

Echoes back a single string line from an embedded CSV that contains airport related metadata.

`AIRPORTS` returns a more production like data with larger message sizes compared to `ECHO`. 

Example lines:

| id   | ident | type          | name                                        | latitude_deg   | longitude_deg      | elevation_ft | continent | iso_country | iso_region | municipality | scheduled_service | gps_code | iata_code | local_code | home_link                   | wikipedia_link                                                            | keywords                                                                        |
|------|-------|---------------|---------------------------------------------|----------------|--------------------|--------------|-----------|-------------|------------|--------------|-------------------|----------|-----------|------------|-----------------------------|---------------------------------------------------------------------------|---------------------------------------------------------------------------------|
| 6523 | 00A   | heliport      | Total Rf Heliport                           | 40.07080078125 | -74.93360137939453 | 11           | NA        | US          | US-PA      | Bensalem     | no                | 00A      |           | 00A        |                             |                                                                           |                                                                                 |
| 4296 | LHBP  | large_airport | Budapest Liszt Ferenc International Airport | 47.42976       | 19.261093          | 495          | EU        | HU          | HU-PE      | Budapest     | yes               | LHBP     | BUD       |            | http://www.bud.hu/english   | https://en.wikipedia.org/wiki/Budapest_Ferenc_Liszt_International_Airport | Ferihegyi nemzetk√∂zi rep√ºl≈ët√©r, Budapest Liszt Ferenc international Airport |
| 3622 | KJFK  | large_airport | John F Kennedy International Airport        | 40.639801      | -73.7789           | 13           | NA        | US          | US-NY      | New York     | yes               | KJFK     | JFK       | JFK        | https://www.jfkairport.com/ | https://en.wikipedia.org/wiki/John_F._Kennedy_International_Airport       | Manhattan, New York City, NYC, Idlewild                                         |


### Subsequent requests

Subsequent request URLs can be set using the `REQUESTS` environment variable. Multiple URLs can be set and must be separated by `space`. A `count` must also be set for each URL using the following syntax: `URL#count`.

### Apache Kafka

Allspark can be used as an Apache Kafka consumer or producer.
A single instance can work as both at the same time.

For producing messages to the kafka cluster, the following workloads are supported:
- `ECHO`
- `AIRPORTS`

Available options:
- `KAFKA_BOOTSTRAP_SERVER`

  The kafka bootstrap server where the cluster can be reached.

  Required for both consumers and producers.

  e.g.
  ```
    name: KAFKA_BOOTSTRAP_SERVER
    value: "localhost:9092"
  ```
- `KAFKA_CONSUMER`

  Starts a kafka consumer for the topic passed in as value.

  e.g.
  ```
    name: KAFKA_CONSUMER
    value: "example-topic"
  ```
- `KAFKA_CONSUMER_GROUP`

  Sets the consumer group id for the kafka consumer.

  If not set it gets defaulted to `allspark-consumer-group`.

  e.g.
  ```
    name: KAFKA_CONSUMER_GROUP
    value: "example-group"
  ```
- `KAFKA_PRODUCER`

  Starts a kafka producer for the topic passed in as value.

  e.g.
  ```
    name: KAFKA_CONSUMER
    value: "example-topic"
  ```


### Example deployment

```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: ratings-v1
  labels:
    app: ratings
    version: v1
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: ratings
        version: v1
    spec:
      containers:
      - name: ratings
        image: banzaicloud/allspark:0.0.2
        imagePullPolicy: Always
        ports:
        - containerPort: 9080
        env:
        - name: SERVER_LISTENADDRESS
          value: 0.0.0.0:9080
        - name: WORKLOAD
          value: Echo
        - name: ECHO_STR
          value: "ratings service response"
        - name: REQUESTS
          value: "http://analytics:8080/#1"
        - name: SQL_DSN
          value: "postgresql://username:password@postgres:5432/postgres?sslmode=allow"
        - name: SQL_QUERY
          value: "SELECT * FROM pg_tables"
        - name: SQL_QUERY_REPEAT_COUNT
          value: "2"
        - name: SQL_QUERY_REPEAT_COUNT_MAX
          value: "10"
---
apiVersion: v1
kind: Service
metadata:
  name: ratings
  labels:
    app: ratings
    service: ratings
spec:
  ports:
  - port: 9080
    name: http
  selector:
    app: ratings
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: analytics-v1
  labels:
    app: analytics
    version: v1
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: analytics
        version: v1
    spec:
      containers:
      - name: analytics
        image: banzaicloud/allspark:0.0.2
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
        env:
        - name: WORKLOAD
          value: PI
        - name: PI_COUNT
          value: "20000"
---
apiVersion: v1
kind: Service
metadata:
  name: analytics
  labels:
    app: analytics
    service: analytics
spec:
  ports:
  - port: 8080
    name: http
  selector:
    app: analytics
```

