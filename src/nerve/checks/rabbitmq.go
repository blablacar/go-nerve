package checks

import (
	log "github.com/Sirupsen/logrus"
	"github.com/streadway/amqp"
	"strconv"
	"bytes"
)

const CHECK_RABBITMQ_TYPE string = "RABBITMQ"

type rabbitmqCheck struct {
	Check
	Username string
	Password string
	VHost string
	Queue string
}

//Initialize 
func(rc *rabbitmqCheck) Initialize() error {
	//Default value pushed here
	rc.Status = StatusUnknown
	rc.Host = "localhost"
	rc.Port = 5672
	rc.IP = "127.0.0.1"
	rc.VHost = ""
	rc.Queue = "nerve"
	rc.Username = "nerve"
	rc.Password = "nerve"
	rc.ConnectTimeout = 10
	rc.DisconnectTimeout = 10
	rc._type = CHECK_RABBITMQ_TYPE
	return nil
}

func(rc *rabbitmqCheck) SetRabbitMQConfiguration(Username string, Password string, Queue string, VHost string) {
	if Username != "" {
		rc.Username = Username
	}
	if Password != "" {
		rc.Password = Password
	}
	if Queue != "" {
		rc.Queue = Queue
	}
	if VHost != "" {
		if VHost == "/" {
			rc.VHost = ""
		}else {
			rc.VHost = VHost
		}
	}
}

//Verify that the given host or ip / port is healthy
func(rc *rabbitmqCheck) DoCheck() (status int, err error) {
	//First connect
	conn, err := amqp.Dial("amqp://"+rc.Username+":"+rc.Password+"@"+rc.IP+":"+strconv.Itoa(rc.Port)+"/"+rc.VHost)
	if err != nil {
		log.WithError(err).Warn("Unable to Connect to RabbitMQ")
		return StatusKO, err
	}
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		log.WithError(err).Warn("Unable to Open Channel in RabbitMQ Connection")
		return StatusKO, err
	}
	queue, err := channel.QueueDeclare(rc.Queue,false,false,false,false,nil)
	if err != nil {
		conn.Close()
		log.WithError(err).Warn("Unable to Declare Queue[",rc.Queue,"] in RabbitMQ Connection")
		return StatusKO, err
	}
	err = channel.Publish("",queue.Name,true,false,amqp.Publishing {
		ContentType: "text/plain",
		Body: []byte("nerve"),
		})
	if err != nil {
		conn.Close()
		log.WithError(err).Warn("Unable to Publish message in RabbitMQ Connection")
		return StatusKO, err
	}
	delivery, ok, err := channel.Get(queue.Name,true)
	if !ok {
		conn.Close()
		if err != nil {
			log.WithError(err).Warn("Error getting message from the RabbitMQ Queue [",rc.Queue,"]")
			return StatusKO, err
		}else {
			log.Warn("Error no message in the RabbitMQ Queue [",rc.Queue,"]")
			return StatusKO, err
		}
	}
	if delivery.Body != nil && bytes.Compare(delivery.Body,[]byte("nerve")) != 0 {
		conn.Close()
		log.WithField("Body",delivery.Body).WithField("queue",rc.Queue).Warn("Error Invalid message Body expect \"nerve\"")
		return StatusKO, err
	}

	return StatusOK, nil
}

func(x *rabbitmqCheck) GetType() string {
	return x._type
}

func (x *rabbitmqCheck) SetBaseConfiguration(IP string, Host string, Port int, ConnectTimeout int, ipv6 bool) {
	x.IP = IP
	x.Host = Host
	x.Port = Port
	x.ConnectTimeout = ConnectTimeout
	x.IPv6 = ipv6
}
