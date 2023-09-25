package random

var (
	letterRunes       = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
	alphanumericRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
)

// String returns a random string of alphanumeric characters. The first
// character will be a letter. If you need the first character to include
// digits, generate a longer string and trim the first character.
func String(length int) string {
	result := make([]rune, length)
	// First character must be a letter
	result[0] = letterRunes[rng.Intn(len(letterRunes))]

	// Rest of the characters can be alphanumeric
	for i := 1; i < length; i++ {
		result[i] = alphanumericRunes[rng.Intn(len(alphanumericRunes))]
	}

	return string(result)
}
