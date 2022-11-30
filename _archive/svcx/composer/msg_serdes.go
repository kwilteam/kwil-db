package composer

type msg_serdes struct {
	//Serdes
}

func (_ *msg_serdes) Serialize(m *Message) ([]byte, []byte, error) {
	panic("implement me")
}

func (_ *msg_serdes) Deserialize(key, value []byte) (*Message, error) {
	panic("implement me")
}
