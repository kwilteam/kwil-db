package mapper

// This package in general will need to be more comprehensive.

type TopicMapper interface {
	GetTopic(msg_ctx MessageContext) string
}

type DbMapper interface {
	GetDbConnection(msg_ctx MessageContext) string
}

type MessageContext interface {
}
