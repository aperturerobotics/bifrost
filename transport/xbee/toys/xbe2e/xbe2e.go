package main

import (
	"context"
	"os"

	"github.com/aperturerobotics/bifrost/transport/xbee/xbserial"
	"github.com/sirupsen/logrus"
	"github.com/tarm/serial"
	"github.com/urfave/cli"
)

var cmdArgs struct {
	// Device1 is the config for the first serial device.
	Device1 serial.Config
	// Device1 is the config for the second serial device.
	Device2 serial.Config
}

// initially we will just try sending data between
// later, test with the full controllerbus stack

func main() {
	app := cli.NewApp()
	app.Name = "xbe2e"
	app.Usage = "end to end test between two xbees"
	app.Action = runE2E
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "device-path-1",
			Usage:       "open device 1 at `PATH`",
			Destination: &cmdArgs.Device1.Name,
			Value:       "/dev/ttyUSB0",
		},
		cli.IntFlag{
			Name:        "device-baud-1",
			Usage:       "baudrate for the first device",
			Destination: &cmdArgs.Device1.Baud,
			Value:       115200,
		},
		cli.StringFlag{
			Name:        "device-path-2",
			Usage:       "open device 2 at `PATH`",
			Destination: &cmdArgs.Device2.Name,
			Value:       "/dev/ttyUSB1",
		},
		cli.IntFlag{
			Name:        "device-baud-2",
			Usage:       "baudrate for the second device",
			Destination: &cmdArgs.Device2.Baud,
			Value:       115200,
		},
	}
	if err := app.Run(os.Args); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	}
}

func runE2E(cctx *cli.Context) error {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	le.Info("starting test: opening device 1")
	sp1, err := serial.OpenPort(&cmdArgs.Device1)
	if err != nil {
		return err
	}
	defer sp1.Close()

	le.Info("starting test: opening device 2")
	sp2, err := serial.OpenPort(&cmdArgs.Device2)
	if err != nil {
		return err
	}
	defer sp2.Close()

	errCh := make(chan error, 3)

	le.Info("starting xbserial 1")
	s1 := xbserial.NewXBeeSerial(le.WithField("xbee", 1), sp1)
	go func() {
		errCh <- s1.ReadPump()
	}()

	le.Info("starting xbserial 2")
	s2 := xbserial.NewXBeeSerial(le.WithField("xbee", 2), sp2)
	go func() {
		errCh <- s2.ReadPump()
	}()

	// Local addresses
	le.Debug("reading local address 1")
	l1, err := s1.ReadLocalAddress(ctx)
	if err != nil {
		return err
	}
	le.Infof("xbee 1 address: %s %x", l1, l1)

	le.Debug("reading local address 2")
	l2, err := s2.ReadLocalAddress(ctx)
	if err != nil {
		return err
	}
	le.Infof("xbee 2 address: %s %x", l2, l2)

	dataCh := make(chan []byte, 5)
	s1.SetDataHandler(func(data []byte, addr xbserial.XBeeAddr) {
		if addr != l2 {
			le.Fatalf("expected incoming address to be %v, got %v", l2, addr)
		} else {
			le.Infof("s1 rx'd data: %v %s", data, string(data))
		}
		select {
		case dataCh <- data:
		default:
		}
	})

	// Write data 2 -> 1
	le.Debug("trying to transmit data from 2 -> 1")
	if err := s2.TxToAddr(ctx, uint64(l1), 0, 0, 0, 0, 0, 0, 0, []byte("hello world")); err != nil {
		return err
	}

	le.Info("running")
	select {
	case <-dataCh:
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}

	le.Info("conveyed data successfully")
	return nil
}
