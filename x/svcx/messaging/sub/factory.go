package sub

//import (
//	"fmt"
//	"kwil/x/cfgx"
//	"kwil/x/composer/mx"
//)
//
//func NewChannelBroker[T any](config cfgx.Config, serdes mx.Serdes[T]) (ChannelBroker[T], error) {
//	if config == nil {
//		return nil, fmt.Errorf("config is nil")
//	}
//
//	cfg, err := mx.NewReceiverConfig[T](config, serdes)
//	if err != nil {
//		return nil, err
//	}
//
//	return new_channel_broker[T](cfg)
//}
