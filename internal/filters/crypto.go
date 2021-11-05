package filters

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"
	"reflect"
	"strings"
)

func SumMD5(value reflect.Value) (reflect.Value, error) {
	if err := isString(value); err != nil {
		return zero, err
	}
	return checksum(value, md5.New())
}

func SumSHA(value reflect.Value) (reflect.Value, error) {
	return checksum(value, sha1.New())
}

func SumSHA256(value reflect.Value) (reflect.Value, error) {
	return checksum(value, sha256.New())
}

func SumSHA512(value reflect.Value) (reflect.Value, error) {
	return checksum(value, sha512.New())
}

func checksum(value reflect.Value, h hash.Hash) (reflect.Value, error) {
	var r io.Reader
	if f, err := os.Open(value.String()); err == nil {
		r = f
		defer f.Close()
	} else {
		r = strings.NewReader(value.String())
	}
	_, err := io.Copy(h, r)
	if err == nil {
		value.SetString(fmt.Sprintf("%x", h.Sum(nil)))
	}
	return zero, err
}
