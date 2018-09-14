package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/mail"
	"strings"
)

func ComputeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	fmt.Println(string(h.Sum(nil)))
	return hex.EncodeToString(h.Sum(nil))
}

func CheckMAC(message, messageMAC, key string) bool {
	// fmt.Println(message, messageMAC, key)
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	expectedMAC := mac.Sum(nil)

	computedMAC, _ := hex.DecodeString(messageMAC)
	return hmac.Equal(computedMAC, expectedMAC)
}

func validReply(toAddress string) bool {
	e, err := mail.ParseAddress(toAddress)
	if err != nil {
		return false
	}

	if !strings.HasPrefix(e.Address, "reply+") {
		return false
	}

	parts := strings.Split(e.Address, "-")
	// fmt.Println("parts", parts, len(parts))

	if len(parts) < 2 {
		return false
	}

	// fmt.Println("parts", parts)
	replyParts := strings.Split(parts[0], "+")

	if len(replyParts) != 2 {
		return false
	}

	endParts := strings.Split(parts[1], "@")

	if len(endParts) != 2 {
		return false
	}

	return CheckMAC(replyParts[1], endParts[0], "foobar")

}

func main() {
	// msg := "12345"
	// sharedsecret := "foobar"
	// mac := ComputeHmac256(msg, sharedsecret)
	// fmt.Println("reply+12345-" + mac + "@dev.unee-t.com")
	// fmt.Println(CheckMAC(msg, mac, sharedsecret))
	fmt.Println(validReply("reply+12345-88cc307ac43f0e395d7a7389ce89313d9f88ff5df89cb74592859bfe6c1f5e9a@dev.unee-t.com"))
}
