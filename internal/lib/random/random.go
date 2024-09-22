package random

import "math/rand"

func NewRandomString(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	lettersLen := len(letters)

	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(lettersLen)]
	}
	return string(b)
}
