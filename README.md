[![Build Status](https://travis-ci.org/blablacar/go-nerve.png?branch=master)](https://travis-ci.org/blablacar/go-nerve)


# Nerve

Nerve is a utility to tracking the status of services. It run different checks against the service, and can report the status to different system.
Though the api, you can also manually control what is reported (disabled, enabled, forced enabled).
Beyond checking and reporting, nerve can execute different command against the service to control warmup, or prepare the service before announcing it.

At BlaBlaCar, we use a nerve process for each service instance as the control point of it. Reporting the status in Zookeeper, and [Synapse](https://github.com/blablacar/go-synapse) on the other side to control access to the service instance.

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

Very minimal configuration file : 
```
services:
  - port: 80
  
# nerve will assume it's localhost
# nerve will add tcp check
# nerve will add console report
```


Root attributes :

```
apiHost: 127.0.0.1
apiPort: 3454
disableWaitInMilli: 3000 # minimum shutdown time, just to be sure status is reported
services: # more complete description below
  - name: my-api 
    port: 80

```

### Services Config

Each service that nerve will be monitoring is specified in the `services` hash.
This is a configuration hash telling nerve how to monitor the service.
The configuration contains the following options:

* `host` (required): the default host on which to make service checks; you should make this your *public* ip to ensure your service is publically accessible
* `port` (required): the default port for service checks; nerve will report the `ip`:`port` combo via your chosen reporter (if you give a real hostname, it will be translated into an IP)
* `check_interval` (optional): the frequency with which service checks (and report) will be initiated in milliseconds; defaults to `500`
* `reporter` (required): a hash containing all information to report if the service is up or down
* `watcher` (required): a hash containing the configuration to check if the service is up or down

### Reporter Config

* `type` (required): the mechanism used to report up/down information; depending on the reporter you choose, additional parameters may be required. Defaults to `console`
* `weight` (optional): a positive integer weight value which can be used to affect the haproxy backend weighting in synapse.
* `haproxy_server_options` (optional): a string containing any special haproxy server options for this service instance. For example if you wanted to set a service instance as a backup.
* `rise` (optional): how many consecutive checks must pass before the check is reported; defaults to 1
* `fall` (optional): how many consecutive checks must fail before the check is reported; defaults to 1
* `tags` (optional): an array of strings to pass to the reporter.

#### Zookeeper Reporter

If you set your reporter `type` to `"zookeeper"` you should also set these parameters:

* `hosts` (required): a list of the zookeeper hosts comprising the [ensemble](https://zookeeper.apache.org/doc/r3.1.2/zookeeperAdmin.html#sc_zkMulitServerSetup) that nerve will submit registration to
* `path` (required): the path (or [znode](https://zookeeper.apache.org/doc/r3.1.2/zookeeperProgrammers.html#sc_zkDataModel_znodes)) where the registration will be created; nerve will create the [ephemeral node](https://zookeeper.apache.org/doc/r3.1.2/zookeeperProgrammers.html#Ephemeral+Nodes) that is the registration as a child of this path, and the name of this ephemeral node is created with the function CreateProtectedEphemeralSequential from the libray Golang Zookeeper [github.com/samuel/go-zookeeper/zk](https://github.com/samuel/go-zookeeper/)

#### Console Reporter

If you set your reporter `type` to `"console"`, no more parameters are available.
All data will be reported as JSON and printed directly on the std output.

#### File Reporter

If you set your reporter `type` to `"file"` you should also set these parameters:

* `path` (optional): the full path where stand the file to report to (default to '/tmp')
* `filename` (optional): the filename (default to 'nerve.report')
* `mode` (optional): whether to open the file in 'write' mode (override the whole content), or in 'append' mode (default to 'write' mode).

### Watcher Config

* `checks` (required): an array of checks that nerve will perform; if all of the pass, the service will be registered; otherwise, it will be un-registered
* `maintenance_checks` (optional): an array of checks that nerve will perform; if one failed, the service will have the maintenance tag set to true; otherwise, it will be set to false

### Checks

The core of nerve is a set of service checks.
Each service can define a number of checks, and all of them must pass for the service to be registered.
Although the exact parameters passed to each check are different, all take a number of common arguments:

* `type`: (required) the kind of check; you can see available check types (for now 'tcp', 'http' and 'rabbitmq')
* `name`: (optional) a descriptive, human-readable name for the check; it will be auto-generated based on the other parameters if not specified
* `host`: (optional) the host on which the check will be performed; defaults to the `host` of the service to which the check belongs
* `port`: (optional) the port on which the check will be performed; like `host`, it defaults to the `port` of the service
* `timeout`: (optional) maximum time the check can take; defaults to `100ms`

#### TCP Check

If you set your check `type` to `"tcp"`, no more parameters are available.

#### HTTP Check

If you set your check `type` to `"http"` you should also set these parameters:

* `uri` (required): the URI to check

#### HTTP Proxy Check

If you set your check `type` to `"httpproxy"` you should also set these parameters:

* `urls` (required): an array of string containing the full url to tests
* `port` (required): the proxy port to use
* `host` (required): the proxy host to use
* `user` (optional): the proxy username
* `password` (optional): the proxy password

#### RabbitMQ Check

If you set your check `type` to `"rabbitmq"` you should also set these parameters:

* `user` (optional): the user to connect to rabbitmq (default to 'nerve')
* `password` (optional): the password to connect to rabbitmq (default to 'nerve')
* `vhost` (optional): the vhost to check (default to /)
* `queue` (optional): the queue in which get test message (default to 'nerve') 

#### Mysql Check

If you set your check `type` to `"mysql"` you should also set these parameters:

* `user` (optional): the user to connect to mysql (default to 'nerve')
* `password` (optional): the password to connect to msqla (default to 'nerve')
* `sql_request` (optional): the SQL Request used to check the Mysql availability (default to "SELECT 1")

#### Zookeeper Flag Check

At BlaBlaCar, we use this Check as a maintenance check. Typically if a defined flag exist in Zookeeper, then the check fail, and the reporter report a failed service. If you want to use it, put the `type` to `"zkflag"`, and you should also set these parameters:

* `hosts` (required): an array of string of ZK nodes
* `path` (required): the key to verify. If it exists, then the check fail

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request
