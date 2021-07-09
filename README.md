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

#### Echo

Simply echos back a set string

Available options:

- EHCO_STR - string to echo back

### Subsequent requests

Subsequent request URLs can be set using the `REQUESTS` environment variable. Multiple URLs can be set and must be separated by `space`. A `count` must also be set for each URL using the following syntax: `URL#count`.

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
