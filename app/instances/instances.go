package instances

import (
	"flag"
	"fmt"
	. "github.com/codecrafters-io/redis-starter-go/app/internal"
	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"math/rand"
	"net"
	"os"
	"strconv"
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

func initProperties() Properties {
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

type Instance struct {
	props    Properties
	replicas []net.Conn
}

func New() Instance {
	props := initProperties()
	i := Instance{props: props, replicas: nil}
	if props.IsMaster() {
		i.replicas = []net.Conn{}
	}
	return i
}

func (i *Instance) Props() Properties {
	return i.props
}

func (i *Instance) ConnectToMaster() {
	conn, _ := net.Dial("tcp", i.props.masterAddress)
	fmt.Printf("[%s] Attempting handshake ... local %s, remote %s\n", i.props.role, conn.LocalAddr(), conn.RemoteAddr())

	responseBuf := make([]byte, 1024)

	WriteString(conn, resp.Array("ping"))
	_, err := conn.Read(responseBuf)
	if err != nil {
		fmt.Println("Error connecting to master 1/3")
		return
	}

	port := strconv.Itoa(int(i.props.port))
	replconfArgs := [][]string{
		{"REPLCONF", "listening-port", port},
		{"REPLCONF", "capa", "psync2"},
	}
	for _, argSet := range replconfArgs {
		WriteString(conn, resp.Array(argSet...))
		_, err = conn.Read(responseBuf)
		if err != nil {
			fmt.Println("Error connecting to master 2/3")
			return
		}
	}

	WriteString(conn, resp.Array("PSYNC", "?", "-1"))
	_, err = conn.Read(responseBuf)
	if err != nil {
		fmt.Println("Error connecting to master 3/3")
		return
	}

	fmt.Printf("[%s] Handshake with master instance successful\n", i.props.Role())
}

func (i *Instance) LinkReplica(replicaConn net.Conn) {
	fmt.Printf("Attaching replica to master, local %s, remote %s\n", replicaConn.LocalAddr(), replicaConn.RemoteAddr())
	i.replicas = append(i.replicas, replicaConn)
}

func (i *Instance) PropagateCommand(b []byte) {
	for _, conn := range i.replicas {
		_, err := conn.Write(b)
		if err != nil {
			fmt.Println("err when sending command", err)
		}
	}
}

func genReplId() string {
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 40)
	for i := range b {
		b[i] = charSet[seed.Intn(len(charSet)-1)]
	}

	return string(b)
}
