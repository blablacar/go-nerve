package nerve

import (
	"github.com/n0rad/go-erlog/errs"
	"github.com/streadway/amqp"
	"strconv"
)

const checkMessage = "nerve"

type CheckAmqp struct {
	CheckCommon
	Vhost    string
	Queue    string
	Username string
	Password string

	url string
}

func NewCheckAmqp() *CheckAmqp {
	return &CheckAmqp{
		Vhost: "/",
		Queue: "nerve",
	}
}

func (x *CheckAmqp) Init(conf *Service) error {
	x.url = "amqp://" + x.Username + ":" + x.Password + "@" + x.Host + ":" + strconv.Itoa(x.Port) + "/" + x.Vhost
	x.fields = x.fields.WithField("url", x.url).WithField("queue", x.Queue)
	return nil
}

func (x *CheckAmqp) Check() error {
	conn, err := amqp.Dial(x.url)
	if err != nil {
		return errs.WithEF(err, x.fields, "Unable to connect to amqp server")
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		return errs.WithEF(err, x.fields, "Unable to open channel")
	}
	defer channel.Close()

	if x.Queue == "" {
		return nil
	}

	queue, err := channel.QueueDeclare(x.Queue, false, false, false, false, nil)
	if err != nil {
		return errs.WithEF(err, x.fields, "Unable to declare queue")
	}

	if err := channel.Publish("", queue.Name, true, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(checkMessage),
	}); err != nil {
		return errs.WithEF(err, x.fields, "Failed to publish message")
	}
	delivery, ok, err := channel.Get(queue.Name, true)
	if !ok || err != nil {
		msg := "Failed to get test message"
		if err == nil {
			msg += ". No message in queue"
		}
		return errs.WithEF(err, x.fields, msg)
	}
	if delivery.Body == nil || string(delivery.Body) != checkMessage {
		return errs.WithEF(err, x.fields.WithField("received", delivery.Body), "Body received do not match sent")
	}
	delivery.Ack(false)
	return nil
}
