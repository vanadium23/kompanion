package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// PartialMD5 returns the MD5 hash of the first 10KB of the file
// See at https://github.com/koreader/koreader/blob/03aa96dc7dc25c8d58977f0165630af7e4514891/frontend/util.lua#L1055
func PartialMD5(filepath string) (string, error) {
	if filepath == "" {
		return "", fmt.Errorf("file path is empty")
	}

	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	step := int64(1024)
	size := 1024
	hash := md5.New()

	for i := -1; i <= 10; i++ {
		offset := int64(0)
		if i >= 0 {
			offset = step << (2 * i)
		}

		_, err := file.Seek(int64(offset), io.SeekStart)
		if err != nil {
			return "", err
		}

		buf := make([]byte, size)
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return "", err
		}
		if n > 0 {
			hash.Write(buf[:n])
		} else {
			break
		}
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
