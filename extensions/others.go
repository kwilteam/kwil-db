//go:build ext_test

package extensions

import (
	_ "github.com/kwilteam/kwil-db/extensions/listeners/spammer"
	_ "github.com/kwilteam/kwil-db/extensions/resolutions/spam"
)
