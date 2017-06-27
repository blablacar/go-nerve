package nerve

import (
	"bytes"
	"sync"
	"text/template"

	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"github.com/streadway/amqp"
)

const checkMessage = "nerve"

type CheckAmqp struct {
	CheckCommon
	Datasource string
	Vhost      string
	Queue      string
	Username   string
	Password   string

	templatedDatasource string
}

func NewCheckAmqp() *CheckAmqp {
	return &CheckAmqp{
		Datasource: "amqp://{{.Username}}:{{.Password}}@{{.Host}}:{{.Port}}/{{.Vhost}}",
		Queue:      "nerve",
	}
}

func (x *CheckAmqp) Run(statusChange chan Check, stop <-chan struct{}, doneWait *sync.WaitGroup) {
	x.CommonRun(x, statusChange, stop, doneWait)
}

func (x *CheckAmqp) Init(s *Service) error {
	if err := x.CheckCommon.CommonInit(s); err != nil {
		return err
	}

	template, err := template.New("datasource").Parse(x.Datasource)
	if err != nil {
		return errs.WithEF(err, x.fields, "Failed to parse datasource template")
	}

	var buff bytes.Buffer
	if err := template.Execute(&buff, x); err != nil {
		return errs.WithEF(err, x.fields, "Datasource templating failed")
	}
	x.templatedDatasource = buff.String()
	logs.WithF(x.fields.WithField("datasource", x.templatedDatasource)).Debug("datasource templated")
	return nil
}

func (x *CheckAmqp) Check() error {
	conn, err := amqp.Dial(x.templatedDatasource)
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
