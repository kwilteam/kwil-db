package mx

import (
	"crypto/tls"
	"fmt"
	"github.com/google/uuid"
	"kwil/x/cfgx"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type ClientConfig struct {
	Brokers        []string
	User           string
	Pwd            string
	Linger         time.Duration
	Client_id      string
	Buffer         int
	Dialer         *tls.Dialer
	Group          string
	DefaultTopic   string
	ConsumerTopics []string
	MaxPollRecords int
}

func NewEmitterClientConfig(config cfgx.Config) (cfg *ClientConfig, err error) {
	buffer := int(config.Int32("out_buffer_size", 10))
	if buffer < 0 {
		err = fmt.Errorf("out_buffer_size cannot be a negative #")
		return
	}

	lingerStr := config.GetString("linger-ms", "50")
	linger, err := strconv.Atoi(lingerStr)
	if err != nil || linger < 0 {
		return nil, fmt.Errorf("invalid linger.ms")
	}

	user, pwd, dialer, brokers, client_id, err := getCommonConfig(config)
	if err != nil {
		return
	}

	id := config.String("emitter.id")
	if id == "" {
		client_id = "_emitter__" + uuid.New().String()
	}

	return &ClientConfig{
		Brokers:      brokers,
		User:         user,
		Pwd:          pwd,
		Linger:       time.Duration(linger) * time.Millisecond,
		Client_id:    client_id,
		Buffer:       buffer,
		Dialer:       dialer,
		DefaultTopic: config.String("default-topic"),
	}, nil
}

func NewReceiverConfig(config cfgx.Config) (cfg *ClientConfig, err error) {
	topics := strings.Split(config.String("default-topic"), ",")
	if len(topics) == 0 {
		topic := config.String("default-topic")
		if topic == "" {
			err = fmt.Errorf("no topic foudn in config")
			return
		}
		topics = append(topics, topic)
	}

	max_poll_str := config.GetString("max-poll-records", "100")
	max_poll, err := strconv.Atoi(max_poll_str)
	if err != nil || max_poll < 0 {
		return nil, fmt.Errorf("invalid max-poll-records")
	}

	user, pwd, dialer, brokers, client_id, err := getCommonConfig(config)
	if err != nil {
		return
	}

	return &ClientConfig{
		Brokers:        brokers,
		User:           user,
		Pwd:            pwd,
		Client_id:      client_id,
		Dialer:         dialer,
		ConsumerTopics: topics,
		MaxPollRecords: max_poll,
		Group:          config.String("group-id"),
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
