package blockcrypt

import (
	"bytes"
	"crypto/rand"
	"errors"
	"testing"
)

func TestBlockCrypt(t *testing.T) {
	var pass [32]byte
	randomize := func(dat []byte) {
		if _, err := rand.Read(dat); err != nil {
			t.Fatal(err.Error())
		}
	}
	randomize(pass[:])
	for i := BlockCrypt_BlockCrypt_AES256; i <= BlockCrypt_BlockCrypt_MAX; i++ {
		if err := func() error {
			c, err := BuildBlockCrypt(i, pass[:])
			if err != nil {
				return err
			}
			var data [128]byte
			randomize(data[:])
			dataBefore := make([]byte, len(data))
			copy(dataBefore, data[:])
			c.Encrypt(data[:], data[:])
			if bytes.Compare(data[:], dataBefore) == 0 {
				return errors.New("data was identical after encrypt")
			}
			c.Decrypt(data[:], data[:])
			if bytes.Compare(data[:], dataBefore) != 0 {
				return errors.New("data was not identical after decrypt")
			}
			return nil
		}(); err != nil {
			t.Fatalf("block crypt %v: %v", i.String(), err)
		}
	}
}
