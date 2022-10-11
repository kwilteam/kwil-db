package mx

import (
	"crypto/tls"
	"fmt"
	"github.com/google/uuid"
	"kwil/x/cfgx"
	"net"
	"os"
	"strconv"
	"time"
)

type ClientConfig[T any] struct {
	Brokers   []string
	User      string
	Pwd       string
	Linger    time.Duration
	Client_id string
	Buffer    int
	Serdes    Serdes[T]
	Dialer    *tls.Dialer
	Group     string
}

func (e *ClientConfig[T]) ToString() string {
	panic("implement me")
}

func NewEmitterConfig[T any](config cfgx.Config, serdes Serdes[T]) (cfg *ClientConfig[T], err error) {
	buffer := int(config.Int32("out_buffer_size", 10))
	if buffer < 0 {
		err = fmt.Errorf("out_buffer_size cannot be a negative #")
		return
	}

	user, pwd, dialer, brokers, client_id, err := getCommonConfig(config)
	if err != nil {
		return
	}

	lingerStr := config.GetString("linger-ms", "50")
	linger, err := strconv.Atoi(lingerStr)
	if err != nil || linger < 0 {
		return nil, fmt.Errorf("invalid linger.ms")
	}

	return &ClientConfig[T]{
		Brokers:   brokers,
		User:      user,
		Pwd:       pwd,
		Linger:    time.Duration(linger) * time.Millisecond,
		Client_id: client_id,
		Buffer:    buffer,
		Serdes:    serdes,
		Dialer:    dialer,
	}, nil
}

func NewReceiverConfig[T any](config cfgx.Config, serdes Serdes[T]) (cfg *ClientConfig[T], err error) {
	user, pwd, dialer, brokers, client_id, err := getCommonConfig(config)
	if err != nil {
		return
	}

	return &ClientConfig[T]{
		Brokers:   brokers,
		User:      user,
		Pwd:       pwd,
		Client_id: client_id,
		Serdes:    serdes,
		Dialer:    dialer,
	}, nil
}

func getCommonConfig(config cfgx.Config) (user string, pwd string, dialer *tls.Dialer, brokers []string, client_id string, err error) {
	client_id = config.String("client.id")
	if client_id == "" {
		h, _ := os.Hostname()
		client_id = h + "_" + uuid.New().String()
	}

	config = config.Select("broker-settings")
	brokers = config.GetStringSlice("bootstrap.servers", ",", nil)
	if len(brokers) == 0 {
		err = fmt.Errorf("bootstrap.servers is empty")
		return
	}

	user = config.String("username")
	pwd = config.String("password")

	dialer = &tls.Dialer{NetDialer: &net.Dialer{Timeout: 10 * time.Second}}

	return
}
