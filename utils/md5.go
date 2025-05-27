package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
)

func StringMD5(data string) string {
	hash := md5.New()
	//不需要手动将字符串变为字节切片
	io.WriteString(hash, data)
	key := hex.EncodeToString(hash.Sum(nil))
	return key
}
