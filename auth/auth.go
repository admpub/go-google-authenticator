package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"time"
)

var (
	Debug = false
)

type Authenticator struct {
	Interval  int
	SecretKey []byte
	Hash      hash.Hash
}

/*
Get a new Authenticator.
from a string of the ToTP salt,
and a bool of whether to use sha256 (default is sha1)

Also defaults to 30 second interval
*/
func New(salt string, twofiftysix bool) Authenticator {
	var h hash.Hash
	if twofiftysix {
		h = sha256.New()
	} else {
		h = sha1.New()
	}

	return Authenticator{
		Interval:  30,
		Hash:      h,
		SecretKey: bytes.NewBufferString(salt).Bytes(),
	}
}

/*
Construct an new Hmac hash with the secret and hashlib (sha1 or sha256)
*/
func (a Authenticator) Hmac() hash.Hash {
	return hmac.New(func() hash.Hash { return a.Hash }, a.SecretKey)
}

/*
Convenience to get the token for the present time
*/
func (a Authenticator) GetCodeCurrent() (int, int64, error) {
	return a.GetCode(0)
}

/*
Generate the Time-based One Time Passcode.

c = -1 :: the previous code
c = 0  :: the current code
c = 1  :: the next code

the returns are: code, seconds to expire, error
*/
func (a Authenticator) GetCode(c int) (int, int64, error) {
	now := time.Now().Unix()
	t_chunk := (now / int64(a.Interval)) + int64(c)

	buf_in := bytes.Buffer{}
	err := binary.Write(&buf_in, binary.LittleEndian, int32(t_chunk))
	if err != nil {
		return 0, 0, err
	}

	h := a.Hmac()
	h.Write(buf_in.Bytes())
	sum := h.Sum(nil)
	offset := sum[len(sum)-1] & 0xF
	code_sect := sum[offset : offset+4]
	if Debug {
		fmt.Printf("sum:\t\t%t\n", sum)
		fmt.Printf("last:\t\t%t\n", sum[len(sum)-1])
		fmt.Printf("offset:\t\t%t\n", offset)
		fmt.Printf("code_sect:\t%t\n", code_sect)
		fmt.Printf("code_sect:\t%#v\n", code_sect)
	}
	var code int32
	buf_out := bytes.NewBuffer(code_sect)
	err = binary.Read(buf_out, binary.LittleEndian, &code)
	if err != nil {
		return 0, 0, err
	}
	if Debug {
		fmt.Printf("unpacked code:\t%#v\n", code)
		fmt.Printf("unpacked code:\t%b\n", code)
	}
	code = code & 0x7FFFFFFF
	if Debug {
		fmt.Printf("sig bit:\t%#v\n", code)
		fmt.Printf("sig bit:\t%b\n", code)
	}
	code = code % 1000000
	if Debug {
		fmt.Printf("mod1000000:\t%#v\n", code)
		fmt.Printf("mod1000000:\t%b\n", code)
	}
	// need to ensure this is padded to always be 6 long
	if code < 100000 {
	}

	i := int64(a.Interval)
	x := (((now + i) / i) * i) - now
	if Debug {
		fmt.Printf("expires:\t%d\n", x)
	}
	return int(code), x, nil
}
