package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

func ComputeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	fmt.Println(string(h.Sum(nil)))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func CheckMAC(message, messageMAC, key string) bool {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	expectedMAC := mac.Sum(nil)

	computedMAC, err := base64.StdEncoding.DecodeString(messageMAC)
	if err != nil {
		panic(err)
	}
	return hmac.Equal(computedMAC, expectedMAC)
}

func main() {
	msg := "12345"
	sharedsecret := "foobar"
	mac := ComputeHmac256(msg, sharedsecret)
	fmt.Println("reply+12345-" + mac + "@dev.unee-t.com")
	fmt.Println(CheckMAC(msg, mac, sharedsecret))
}
