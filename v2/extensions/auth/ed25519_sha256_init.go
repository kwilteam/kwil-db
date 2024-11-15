//go:build auth_ed25519_sha256 || ext_test

package auth

func init() {
	err := RegisterAuthenticator(ModAdd, Ed25519Sha256Auth, Ed22519Sha256Authenticator{})
	if err != nil {
		panic(err)
	}
}
