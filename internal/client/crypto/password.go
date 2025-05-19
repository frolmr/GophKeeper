package crypto

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
)

const (
	lowerChars   = "abcdefghijklmnopqrstuvwxyz"
	upperChars   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitChars   = "0123456789"
	specialChars = "!@#$%^&*()-_=+,.?/:;{}[]~"
	alphabet     = lowerChars + upperChars + digitChars + specialChars

	masterPasswordLength = 16
)

// NOTE: now MasterPassword is generated randomly
// might be changed to be generated using PBKDF2 or Argon2 using password
// but this is less secure, as loss of the password may lead to MasterKey loss
func (cs *CryptoService) GenerateStrongPassword() (string, error) {
	requiredChars := []string{
		lowerChars,
		upperChars,
		digitChars,
		specialChars,
	}

	var password strings.Builder

	for _, chars := range requiredChars {
		char, err := getRandomChar(chars)
		if err != nil {
			return "", err
		}
		password.WriteByte(char)
	}

	for i := len(requiredChars); i < masterPasswordLength; i++ {
		char, err := getRandomChar(alphabet)
		if err != nil {
			return "", err
		}
		password.WriteByte(char)
	}

	shuffled := []byte(password.String())
	for i := range shuffled {
		j, err := getRandomInt(i, len(shuffled))
		if err != nil {
			return "", err
		}
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return string(shuffled), nil
}

func getRandomChar(charSet string) (byte, error) {
	if charSet == "" {
		return 0, errors.New("empty character set")
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charSet))))
	if err != nil {
		return 0, err
	}
	return charSet[n.Int64()], nil
}

func getRandomInt(minEl, maxEl int) (int, error) {
	if minEl >= maxEl {
		return 0, errors.New("invalid range: min >= max")
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(maxEl-minEl)))
	if err != nil {
		return 0, err
	}
	return minEl + int(n.Int64()), nil
}
