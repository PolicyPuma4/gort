package generate

import (
	"hash/crc32"
	"strconv"

	u "github.com/google/uuid"
)

func NewCode() (code string, err error) {
	uuid, err := u.NewRandom()
	if err != nil {
		return
	}

	code = strconv.FormatInt((int64)(crc32.ChecksumIEEE([]byte(uuid.String()))), 16)

	return
}
