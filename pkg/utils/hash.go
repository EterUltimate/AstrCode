// AstrCode - Agent orchestration engine for AstrBot
// Copyright (C) 2026 EterUltimate
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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
