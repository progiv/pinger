package main

import (
	"flag"
	"fmt"
	"github.com/progiv/pinger/internal"
	"log"
	"os"
	"time"
)

/*
usage ping [options] <host/ip>
!should report loss and RTT times
!emit requests with a periodic delay
!packet loss and latency
![set TTL and report "time exceeded"]
options:
-4 use ipv4
-6 use ipv6
*/

func usage() {
	fmt.Printf("Usage: %s [Options] Host\nOptions:\n", os.Args[0])
	flag.PrintDefaults()
}

/*
Host     string
	Count    int
	Proto    int
	Timeout  time.Duration
	Interval time.Duration
	Ttl      int
 */
func main() {
	var use4 bool
	config := internal.PingConfig{}

	flag.Usage = usage
	flag.BoolVar(&use4, "4", true, "use IPv4")
	flag.IntVar(&config.Count, "c", 0, "stop after <count> replies")
	flag.DurationVar(&config.Interval, "i", time.Second, "Wait interval between sending packets")
	flag.DurationVar(&config.Timeout, "W", time.Second * 5, "Time to wait for a response")
	flag.IntVar(&config.Ttl, "t", 255, "Set the IP Time to leave")
	flag.Parse()

	remainder := flag.Args()
	if len(remainder) != 1 {
		flag.Usage()
		return
	}

	config.Host = remainder[0]

	fmt.Printf("Sending icmp echo to host %s:\n", config.Host)

	/*err := internal.Ping4(config)
	if err != nil {
		log.Fatal(err)
	}*/
	pinger := internal.NewPinger(config)
	err := pinger.Ping()
	if err != nil {
		log.Fatal(err)
	}
}
