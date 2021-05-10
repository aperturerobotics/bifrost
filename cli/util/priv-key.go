package cliutil

import (
	"bytes"
	"io"
	"os"

	"github.com/aperturerobotics/bifrost/keypem"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// RunGeneratePrivate runs the generate-private util command.
func (a *UtilArgs) RunGeneratePrivate(_ *cli.Context) error {
	npeer, err := peer.NewPeer(nil)
	if err != nil {
		return err
	}

	priv := npeer.GetPrivKey()
	pemd, err := keypem.MarshalPrivKeyPem(priv)
	if err != nil {
		return err
	}
	err = writeIfNotExists(a.FilePath, bytes.NewReader(pemd))
	if err != nil {
		return err
	}
	le := a.GetLogger()
	le.Debugf("generated private key: %s", npeer.GetPeerID().Pretty())
	return nil
}

func writeIfNotExists(outPath string, input io.Reader) error {
	var of *os.File
	var out io.Writer
	if outPath != "" {
		_, err := os.Stat(outPath)
		if !os.IsNotExist(err) {
			return errors.Wrap(os.ErrExist, outPath)
		}
		of, err = os.OpenFile(outPath, os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			return err
		}
		out = of
		defer of.Close()
		if pos, err := of.Seek(0, os.SEEK_END); err != nil || pos != 0 {
			if err == nil {
				// file must have existed
				return errors.Wrap(os.ErrExist, outPath)
			}
			return err
		}
	} else {
		out = os.Stdout
	}
	if _, err := io.Copy(out, input); err != nil {
		return err
	}
	if of != nil {
		return of.Close()
	}
	return nil
}
