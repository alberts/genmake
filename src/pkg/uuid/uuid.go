// Universally Unique IDentifier (UUID).
package uuid

// RFC 4122: A Universally Unique IDentifier (UUID) URN Namespace.

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
)

type Uuid []byte

func Make() Uuid {
	return make(Uuid, 16)
}

// Make Version 4 (random data based) UUID.
func MakeV4() Uuid {
	// V4 UUID is of the form: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	// where x is any hexadecimal digit and y is one of 8, 9, A, or B.
	uuid := Make()
	_, err := rand.Read(uuid)
	if err != nil {
		panic(err)
	}

	// Set the four most significant bits (bits 12 through 15) of the
	// time_hi_and_version field to the 4-bit version number from
	// Section 4.1.3.
	uuid[6] = (uuid[6] & 0xf) | 0x40

	// Set the two most significant bits (bits 6 and 7) of the
	// clock_seq_hi_and_reserved to zero and one, respectively.
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return uuid
}

func Parse(str string) (Uuid, os.Error) {
	if len(str) == 38 {
		if str[0] != '{' || str[37] != '}' {
			return nil, os.EINVAL
		}
		str = str[1:37]
	}
	if len(str) != 36 {
		return nil, os.EINVAL
	}
	uuid := Make()
	j := 0
	for i, c := range str {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return nil, os.EINVAL
			}
			continue
		}
		var v byte
		if c >= 'a' && c <= 'f' {
			v = 10 + byte(c-'a')
		} else if c >= 'A' && c <= 'F' {
			v = 10 + byte(c-'A')
		} else if c >= '0' && c <= '9' {
			v = byte(c - '0')
		} else {
			return nil, os.EINVAL
		}
		if j&0x1 == 0 {
			uuid[j>>1] = v << 4
		} else {
			uuid[j>>1] |= v
		}
		j++
	}
	version := uuid.Version()
	if version < 1 || version > 5 {
		return nil, os.EINVAL
	}
	return uuid, nil
}

func (uuid Uuid) Version() int {
	return int(uuid[6] >> 4)
}

func (uuid Uuid) Equal(other Uuid) bool {
	return bytes.Equal(uuid[0:], other[0:])
}

func (uuid Uuid) String() string {
	if uuid == nil {
		return "nil"
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", []byte(uuid[0:4]), []byte(uuid[4:6]), []byte(uuid[6:8]), []byte(uuid[8:10]), []byte(uuid[10:]))
}

type UuidKey string

func (uuid Uuid) Key() UuidKey {
	return UuidKey([]byte(uuid))
}
