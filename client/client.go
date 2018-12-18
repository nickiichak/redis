package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

func startup() (host, port string, err error) {
	args := os.Args
	//default values
	port = ":9090"
	host = "127.0.0.1"
	err = nil

	for i := 1; i < len(args); i += 2 {
		switch args[i] {
		// -p, --port	: set the port for listening on
		case "-p", "--port":
			var str string
			if args[i+1][0] == ':' {
				str = args[i+1][1:]
			} else {
				str = args[i+1]
			}
			if _, error := strconv.Atoi(str); error == nil {
				port = ":" + str
			} else {
				err = errors.New("ERROR: Wrong port\n")
				return
			}
		// -m, --mode	: enable mirroring data to the drive
		case "-h", "--host":
			host = args[i+1]
		default:
			err = errors.New("ERROR: Unknown command\n")
			return
		}
	}
	return
}


func main() {
	host, port, cmd_err := startup()
	addr := host + port

	//No errors, open port, start redis, start listening
	if cmd_err == nil {
		conn, net_err := net.Dial("tcp", addr)
		if net_err != nil {
			log.Fatalln(net_err)
		} else {
			defer conn.Close()
			fmt.Println("Connected to", addr)
			for {
				// read in input from stdin
				reader := bufio.NewReader(os.Stdin)
				command, _ := reader.ReadString('\n')
				fmt.Fprintf(conn, command)
				message, _ := bufio.NewReader(conn).ReadString('\n')
				fmt.Print(message)
			}
		}

	} else {
		//Some CMD error
		fmt.Println(cmd_err)
	}
}