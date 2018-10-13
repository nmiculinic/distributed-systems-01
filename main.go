package main

import (
	"math"
	"net"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use: "main",
	}
	addr := root.PersistentFlags().String("addr", "127.0.0.1:20120", "address to send/listen to")

	recv := &cobra.Command{
		Use:   "recv",
		Short: "recv data",
	}
	summaryInterval := recv.Flags().Duration("interval", time.Second, "msg interval")

	recv.RunE = func(cmd *cobra.Command, args []string) error {
		ServerAddr, err := net.ResolveUDPAddr("udp", *addr)
		if err != nil {
			return err
		}

		/* Now listen at selected port */
		conn, err := net.ListenUDP("udp", ServerAddr)
		if err != nil {
			return err
		}
		defer conn.Close()
		log.Infoln("Listening on UDP", ServerAddr)

		buf := make([]byte, 1024)
		t := time.NewTicker(*summaryInterval)

		numbers := make(chan int64)
		total := 0
		var minN int64 = math.MaxInt64
		var maxN int64 = math.MinInt64

		go func() {
			for {
				n, addr, err := conn.ReadFromUDP(buf)
				if err != nil {
					log.WithError(err).Errorln("cannot read from udp")
					continue
				}
				x, err := strconv.ParseInt(
					strings.TrimSpace(string(buf[0:n])),
					10,
					64,
				)
				if err != nil {
					log.WithError(err).Errorln("Cannot parse int")
					continue
				}
				log.Debugln("Received ", string(buf[0:n]), " from ", addr)
				numbers <- x
			}
		}()

		for {
			select {
			case <-t.C:
				if total > 0 {
					log.Infof(
						"Recieved %d/%d packages in range [%d-%d]",
						total,
						maxN-minN+1,
						minN,
						maxN,
					)
				} else {
					log.Infoln("No packages recieved")
				}
				total = 0
				minN = math.MaxInt64
				maxN = math.MinInt64
			case x := <-numbers:
				total++
				if x > maxN {
					maxN = x
				}
				if x < minN {
					minN = x
				}
			}
		}
	}
	root.AddCommand(recv)

	send := &cobra.Command{
		Use:   "send",
		Short: "sends data",
		Args:  cobra.NoArgs,
	}

	writeInterval := send.Flags().Duration("interval", 100*time.Millisecond, "msg interval")
	send.RunE = func(cmd *cobra.Command, args []string) error {
		conn, err := net.ListenPacket("udp", ":0")
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		dst, err := net.ResolveUDPAddr("udp", *addr)
		if err != nil {
			log.Fatal(err)
		}

		// The connection can write data to the desired address.
		log.Infof("Sending data to %v every %v", *addr, *writeInterval)
		for i := 0; ; i++ {
			_, err = conn.WriteTo([]byte(strconv.Itoa(i)+"\n"), dst)
			if err != nil {
				log.WithError(err).Errorln("cannot send UDP package")
			}
			time.Sleep(*writeInterval)
		}
	}
	root.AddCommand(send)
	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
