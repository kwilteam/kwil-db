/*
Package auth contains any known Authenticator extensions that may be selected at
build-time for use in kwild. Authenticator extensions are used to expand the
type of signatures that may be verified, and define address derivation for the
public keys of the corresponding type.

Build constraints a.k.a. build tags are used to enable extensions in a kwild
binary. See README.md in the extensions package for more information.
*/
package auth
