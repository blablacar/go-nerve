[![Build Status](https://travis-ci.org/blablacar/go-nerve.png?branch=master)](https://travis-ci.org/blablacar/go-nerve)


# Nerve

Nerve is a utility to tracking the status of services. It run different checks against the service, and can report the status to different system.
Though the api, you can also manually control what is reported (disabled, enabled, forced enabled).
Beyond checking and reporting, nerve can execute different command against the service to control warmup, or prepare the service before announcing it.

At BlaBlaCar, we use a nerve process for each service instance (> 2000) as the control point of it. Reporting the status in Zookeeper, and [Synapse](https://github.com/blablacar/go-synapse) on the other side to control access to the service instance.

## Airbnb

Go-Nerve is a go rewrite of Airbnb's [Nerve](https://github.com/airbnb/nerve) with additional feature to meet our needs.

## Installation

Download the latest version on the [release page](https://github.com/blablacar/go-nerve/releases).

Create a configuration file base on the doc or [examples](https://github.com/blablacar/go-nerve/tree/master/examples).

Run with `./nerve nerve-config.yml`

### Building

Just clone the repository and run `./gomake`

## Configuration

It's a YAML file. You can find examples [here](https://github.com/blablacar/go-nerve/tree/master/examples)

Very minimal configuration file with only one service :
```yaml
services:
  - port: 80
  
# nerve will assume it's localhost
# nerve will add tcp check
# nerve will add console report
```


Root attributes :

```yaml
apiHost: 127.0.0.1
apiPort: 3454
services:
  - port: 80
    ...     # see complete example below
    
    checks: 
      - type: tcp
        ... # see complete example below

    reporters:
      - type: console
        ... # see complete example below
```

### Services Config

Only **port** is mandatory
All commands are an array of string (ex: [/bin/bash, -c, 'echo "salut!" > /tmp/yop'])

```yaml
...
services: 
  - name: '127.0.0.1:80'                          # name of the service. Default to host:port
    host: 127.0.0.1                               # ip of the service
    port: 80                                      # what port is using the service
    preferIpv4: false                             # if using dns instead of ip in checks & dns resolving gives ipv4 & ipv6 results
    weight: 255                                   # service weight when fully available (not on warmup)
    checks: ...
    reporters: ...
    reportReplayInMilli: 1000                     # time wait before replaying failed report
    haproxyServerOptions:                         # a string, to be retrocompatible with airbnb's nerve. Attributes to pushed to synapse's haproxy
    setServiceAsDownOnShutdown: true              #
    labels: {host: r110-srv10}                    # key-value labels to add to the report

    preAvailableCommand:                          # command to run when checks are ok, but service not reported yet (ex: sync queues)
    preAvailableMaxDurationInMilli:               #

    enableCheckStableCommand:                     # command to check that the service is stable with current load during warmup (ex : too many cache missed)
    enableWarmupIntervalInMilli: 2000             # interval between weight going to next value (see below)
    enableWarmupMaxDurationInMilli: 2 * 60 * 1000 # max warmup duration. if reached, warmup is stopped and weight is set as weight value
    disableGracefullyDoneCommand:                 # command to check if the service is gracefully stopped. Usually check if there is still connections
    disableGracefullyDoneIntervalInMilli: 1000    # time wait before relaunching graceful done command
    disableMaxDurationInMilli: 60 * 1000          # maximum service disable time if graceful done is never reached
    disableMinDurationInMilli: 3000               # minimum service disable time, to give at lease some time to users to stop using the service
    noMetrics: false                              # do not include this service in api /metrics report

```

**disable*** are used when calling the api to disable the service.
It's not acting (stopping) directly on the service but more handling what is on the other side of the reporter.
It's useful to consider the service as stoppable when the API /disable reply.

**Enable*** handle the way the service is going up.
As soon as checks are ok, the service is reported with a current weight of **1** and the warmup is triggered.
current weight is increased following a weighted fibonacci suite until reaching **weight** value.
If **enableCheckStableCommand** is set, the command is run at each increase and if returning != 0, current weight restart from 1
until reaching **weight** or **enableWarmupMaxDurationInMilli**.

### Reporter Config

#### Reporter Zookeeper

only hosts and path are mandatory

```yaml
...
services: 
  - ... 
    reporters:
        - type: zookeeper
          hosts: ['127.0.0.1:2181', '127.0.0.1:2182']   # list of zk servers
          path: /services/cassandra/messages            # path to push the key 
          connectionTimeoutInMilli: 2000
          refreshIntervalInMilli: 5 * 60 * 1000         # just in case zookeeper restart from scratch
          exposeOnUnavailable: false                    # insert in zookeeper even if not available. false to be compatible with airbnb's nerve
```

#### Reporter Console

```yaml
...
services: 
  - ... 
    reporters:
        - type: console
```

#### Reporter File

```yaml
...
services: 
  - ... 
    reporters:
        - type: file
          path: /tmp/nerve.report   # this is the default value
          append: false             # Replace file content, or just append to it
```

### Checks

All checks have those attributes in common. None are mandatory

```yaml
...
services: 
  - ... 
    checks:
        - type: XXX
          host: 127.0.0.1            # default is same as service
          port: 80                   # default is same as service
          timeoutInMilli: 1000       # check timeout, and consider as failed
          rise: 3                    # number of check run to consider it's OK
          fall: 3                    # number of check run to consider it's KO
          checkIntervalInMilli: 1000
          ...

```

#### TCP

This check is automatically added if none specified

```yaml
...
services: 
  - ... 
    checks:
        - type: tcp
          ...
```

#### Exec

To implement your own custom check

```yaml
...
services: 
  - ... 
    checks:
        - type: exec
          ...
          command: [/bin/bash, -c, 'echo "salut!" > /tmp/yop']      # mandatory
```

#### HTTP

```yaml
...
services: 
  - ... 
    checks:
        - type: http
          ...
          path: /       # combined with host and port to create the full url
```

Check fail if cannot connect or if status code is in >= 500 && < 600

#### Proxy http

```yaml
...
services: 
  - ... 
    checks:
        - type: proxyhttp
          ...
          proxyHost: 127.0.0.1        # default to service host 
          proxyPort: 80               # default to service port
          proxyUsername:
          proxyPassword:
          urls:                       # list of url to check (ex: ['https://www.google.com/', 'http://status.aws.amazon.com/'])
          failOnAnyUnreachable: false # fail on any unavailable, or fail on all unavailable 
```

#### AMQP

```yaml
...
services: 
  - ... 
    checks:
        - type: amqp
          ...
          datasource: "amqp://{{.Username}}:{{.Password}}@{{.Host}}:{{.Port}}/{{.Vhost}}" # a template. This is the default value
          vhost:                                                                          # default is empty, which in template result in /
          queue: nerve
          username:
          password: 
```

#### SQL

Only `mysql` and `postgres` drivers are embedded. Fill a request if you want more.

```yaml
...
services: 
  - ... 
    checks:
        - type: sql
          ...
          datasource: "{{.Username}}:{{.Password}}@tcp([{{.Host}}]:{{.Port}})/?timeout={{.TimeoutInMilli}}ms"
          driver:  mysql                                                                      
          request: select 1
          username: root
          password: 
```
