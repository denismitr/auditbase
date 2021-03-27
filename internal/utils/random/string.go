package random

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var strMu sync.Mutex
var seededString = rand.New(rand.NewSource(time.Now().UnixNano()))


const alphaNumericCharset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func String(length int) string {
	strMu.Lock()
	defer strMu.Unlock()

	b := make([]byte, length)
	for i := range b {
		b[i] = alphaNumericCharset[seededString.Intn(len(alphaNumericCharset))]
	}
	return string(b)
}

func NumericString(min, max int) string {
	n := Int(min, max)
	return fmt.Sprintf("%d", n)
}
