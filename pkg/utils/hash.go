package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// HashString 计算字符串的 SHA256 哈希
func HashString(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// HashArgs 计算参数的哈希
func HashArgs(args map[string]interface{}) string {
	// 将参数序列化为字符串后哈希
	str := fmt.Sprintf("%v", args)
	return HashString(str)
}

// CombineHashes 组合多个哈希值
func CombineHashes(hashes ...string) string {
	combined := ""
	for _, h := range hashes {
		combined += h
	}
	return HashString(combined)
}
