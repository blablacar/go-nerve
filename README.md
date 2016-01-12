[![Build Status](https://travis-ci.org/blablacar/go-nerve.png?branch=master)](https://travis-ci.org/blablacar/go-nerve)

___/!\ This Software is in Alpha Stage! Don't use it in production, until it's considered as Stable /!\___

# Go-Nerve

Go-Nerve is a utility for tracking the status of machines and services, a Go rewritten work of Airbnb's [Nerve](https://github.com/airbnb/nerve)
It runs locally on the boxes which make up a distributed system, and reports state information to a distributed key-value store.
At BlaBlaCar, we use Zookeeper as our key-value store (same story as Airbnb'a one).
The combination of Nerve and [Synapse](https://github.com/airbnb/synapse) make service discovery in the cloud easy!

## Airbnb ##

Thank you guy to write a so nice piece of software with nerve. But we really want to stop deploying a full ruby stack on our containers ! Our first thoughts were to ask you to rewrote it in C/C++/Java/Go. But our team convince ourself that it was not the best behavior to have at first. So we rewrote it in Go (See more explanations in the Motivation section below).

We want to thanks the huge work made by Airbnb's engineering team. We love you guy's ! Your tools are in the center of our infrastructure at BlaBlaCar. Even if we fork Nerve to rewrote in Go, we will continue to follow your repository, and consider it as the reference. Big Up to YOU! We send you all love and kisses you deserve (and even more).

## Motivation ##

Why rewrote the Airbnb's software ? Firt of all, well, we're not as easy as it seems in Ruby! And, we need to add new features to the tool. 2 choices: learning Ruby, and propose PR, or rewrote in a language we know. We choose the second option. By the way why Go (because we're also easy with Java) ? After compilation, we have a single binary which is easier to deploy on our full container infrastructure! No need to deploy the full ruby stack, nor java one.

We already use [Synapse](https://github.com/airbnb/synapse) to discover remote services.
However, those services needed boilerplate code to register themselves in [Zookeeper](http://zookeeper.apache.org/).
Nerve simplifies underlying services, enables code reuse, and allows us to create a more composable system.
It does so by factoring out the boilerplate into it's own application, which independenly handles monitoring and reporting.

Beyond those benefits, nerve also acts as a general watchdog on systems.
The information it reports can be used to take action from a centralized automation center: action like scaling distributed systems up or down or alerting ops or engineering about downtime.

## Installation ##

### Pre-requisite ###
Verify that you have a decent installation of the Golang compiler, you need one.
Then, we use here the [GOM](https://github.com/mattn/gom) tool to manage dependencies and build the nerve binary. All install information can be found on the github repository:
https://github.com/mattn/gom

Optionnaly, you can also install a GNU Make on your system. It's not needed, but will ease the build and install process.

### Build ###

Clone the repository where you want to have it:

git clone https://github.com/blablacar/go-nerve

Install in _vendor directory all dependencies (for a list take a look at the Gomfile):

	gom install

Then you can build the Nerve Binary:

	gom build nerve/nerve

### Makefile ###
If you have a GNU Make or equivalent on your system, you can also use it to build and install nerve.

	`make dep-install` # Will install all go dependencies into _vendor directory

	`make build` # Will compile nerve binary and push it into local bin/ diretory

	`make install` # Will install nerve binary in the system directory /usr/local/bin (can be overriden at the top of the Makefile)

	`make clean` # Will remove all existing binary in bin/ and remove the dependencies directory _vendor

	`make all` # an alias to make clean dep-install build

## Configuration ##

Go-Nerve depends on a single configuration file, in json format.
It is usually called `nerve.conf.json`.
An example config file is available in `example/nerve.conf.json`.
The config file is composed of four main sections:

* `instance_id`: the name nerve will submit when registering services; makes debugging easier
* `log-level`: The log level (any valid value from DEBUG, INFO, WARN, FATAL)
* `ipv6`: Whether to enable ipv6 management.
* `services`: the hash (from service name to config) of the services nerve will be monitoring

### Services Config ###

Each service that nerve will be monitoring is specified in the `services` hash.
This is a configuration hash telling nerve how to monitor the service.
The configuration contains the following options:

* `host`: the default host on which to make service checks; you should make this your *public* ip to ensure your service is publically accessible
* `port`: the default port for service checks; nerve will report the `ip`:`port` combo via your chosen reporter (if you give a real hostname, it will be translated into an IP)
* `reporter`: a hash containing all information to report if the service is up or down
* `watcher`: a hash containing the configuration to check if the service is up or down

### Reporter Config ###

* `type`: the mechanism used to report up/down information; depending on the reporter you choose, additional parameters may be required. Defaults to `console`
* `weight` (optional): a positive integer weight value which can be used to affect the haproxy backend weighting in synapse.
* `haproxy_server_options` (optional): a string containing any special haproxy server options for this service instance. For example if you wanted to set a service instance as a backup.
* `rise`: (optional) how many consecutive checks must pass before the check is reported; defaults to 1
* `fall`: (optional) how many consecutive checks must fail before the check is reported; defaults to 1

#### Zookeeper Reporter ####

If you set your reporter `type` to `"zookeeper"` you should also set these parameters:

* `hosts`: a list of the zookeeper hosts comprising the [ensemble](https://zookeeper.apache.org/doc/r3.1.2/zookeeperAdmin.html#sc_zkMulitServerSetup) that nerve will submit registration to
* `path`: the path (or [znode](https://zookeeper.apache.org/doc/r3.1.2/zookeeperProgrammers.html#sc_zkDataModel_znodes)) where the registration will be created; nerve will create the [ephemeral node](https://zookeeper.apache.org/doc/r3.1.2/zookeeperProgrammers.html#Ephemeral+Nodes) that is the registration as a child of this path, and the name of this ephemeral node is created with the function CreateProtectedEphemeralSequential from the libray Golang Zookeeper [github.com/samuel/go-zookeeper/zk](https://github.com/samuel/go-zookeeper/)

#### Console Reporter ####

If you set your reporter `type` to `"console"`, no more parameters are available.
All data will be reported as JSON and printed directly on the std output.

#### File Reporter ####

If you set your reporter `type` to `"file"` you should also set these parameters:

* `path`: the full path of the file to report to

### Watcher Config ###

* `checks`: an array of checks that nerve will perform; if all of the pass, the service will be registered; otherwise, it will be un-registered
* `check_interval`: the frequency with which service checks will be initiated in milliseconds; defaults to `500`

### Checks ###

The core of nerve is a set of service checks.
Each service can define a number of checks, and all of them must pass for the service to be registered.
Although the exact parameters passed to each check are different, all take a number of common arguments:

* `type`: (required) the kind of check; you can see available check types (for now 'tcp', 'http' and 'rabbitmq')
* `name`: (optional) a descriptive, human-readable name for the check; it will be auto-generated based on the other parameters if not specified
* `host`: (optional) the host on which the check will be performed; defaults to the `host` of the service to which the check belongs
* `port`: (optional) the port on which the check will be performed; like `host`, it defaults to the `port` of the service
* `timeout`: (optional) maximum time the check can take; defaults to `100ms`

#### TCP Check ####

If you set your check `type` to `"tcp"`, no more parameters are available.

#### HTTP Check ####

If you set your check `type` to `"http"` you should also set these parameters:

* `uri`: the URI to check

#### RabbitMQ Check ####

If you set your check `type` to `"rabbitmq"` you should also set these parameters:

* `username`: the username to connect to rabbitmq
* `password`: the password to connect to rabbitmq
* `vhost`: the vhost to check (default to /)
* `exchange`: the temp exchange name to publish message into (default to 'nerve')
* `bind_name`: the key used to route message from exchange to queue (default to 'nerve')
* `queue`: the queue in which get test message (default to 'nerve') 

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request
