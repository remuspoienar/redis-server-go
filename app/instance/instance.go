package instance

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
)

const (
	charSet     = "aAbBcCdDeEfFgGhHiIjJkKlLmMnNoOpPqQrRsStTuUvVwWxXyYzZ0123456789"
	DefaultPort = 6379
	MASTER      = "master"
	SLAVE       = "slave"
)

type Properties struct {
	role          string
	replId        string
	replOffset    int
	port          uint
	masterAddress string
}

func InitProperties() Properties {
	props := Properties{role: MASTER, replId: genReplId(), replOffset: 0}

	var masterHost string

	flag.UintVar(&props.port, "port", DefaultPort, "Port to run server(positive integer)")
	flag.StringVar(&masterHost, "replicaof", "", "Provide master address to start in replica mode)")
	flag.Parse()
	if masterHost != "" {
		props.role = SLAVE
		var masterPort string

		if len(flag.Args()) > 0 {
			masterPort = flag.Args()[0]
		} else {
			fmt.Println("Invalid or incomplete master address")
			os.Exit(1)
		}
		props.masterAddress = fmt.Sprintf("%s:%s", masterHost, masterPort)
	}

	return props
}

func (props *Properties) Role() string {
	return props.role
}

func (props *Properties) ReplId() string {
	return props.replId
}

func (props *Properties) ReplOffset() int {
	return props.replOffset
}

func (props *Properties) IsMaster() bool {
	return props.role == MASTER
}

func (props *Properties) Port() uint {
	return props.port
}

func (props *Properties) MasterAddress() string {
	return props.masterAddress
}

func (props *Properties) ReplicationInfo() string {
	return fmt.Sprintf(`# Replication
role:%s
master_replid:%s
master_repl_offset:%d
`, props.Role(), props.ReplId(), props.ReplOffset())
}

func genReplId() string {
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 40)
	for i := range b {
		b[i] = charSet[seed.Intn(len(charSet)-1)]
	}

	return string(b)
}