package checks

import "strings"

const (
	StatusOK int = 0
	StatusKO int = 1
	StatusUnknown int = 2
)

type Check struct {
	Status int
	Error error
	_type string
}

type CheckI interface {
	Initialize() error
	DoCheck() (status int, err error)
	GetType() string
}

func CreateCheck(_type string, config []string) (check CheckI, err error) {
        switch (strings.ToUpper(_type)) {
                case CHECK_TCP_TYPE:
                        check = new(tcpCheck)
			check.Initialize()
                case CHECK_HTTP_TYPE:
                        check = new(httpCheck)
			check.Initialize()
                case CHECK_RABBITMQ_TYPE:
                        check = new(rabbitmqCheck)
			check.Initialize()
                default:
                        check = new(tcpCheck)
                        check.Initialize()
        }
        return check, nil
}
