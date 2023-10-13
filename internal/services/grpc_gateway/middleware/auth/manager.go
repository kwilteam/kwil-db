package auth

type token struct {
	ApiKey string
}

type Manager interface {
	IsAllowed(*token) bool
}
