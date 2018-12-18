package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func startup() (port string, disk bool, file_pass string, err error) {
	args := os.Args
	//default values
	port = ":9090"
	disk = false
	file_pass = "./redisDatabase"
	err = nil

	for i := 1; i < len(args); i += 2 {
		switch args[i] {
		// -p, --port	: set the port for listening on
		case "-p", "--port":
			//cutting off ':' if it exist
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
		case "-m", "--mode":
			disk = true
		default:
			err = errors.New("ERROR: Unknown command\n")
			return
		}
	}
	return
}

func redis(clientCh clientChan, saveCh saveChan, disk bool) {
	fmt.Println("redis started")
	if disk {
		fmt.Println("disk mode")
	} else {
		fmt.Println("RAM mode")
	}
	//Redis RAM storage
	var data = make(map[string]string)

	for cmd := range clientCh.input {
		cmd_list := strings.Fields(cmd)
		if len(cmd_list) < 2 || len(cmd_list) > 3 {
			// sending error
			go func() {
				clientCh.output <- ""
				err := errors.New("ERROR: Wrong command\n")
				clientCh.err <- err
			}()
			continue
		}
		switch cmd_list[0] {
		// SET <KEY> <VALUE>
		case "SET":
			if len(cmd_list) < 3 {
				go func() {
					clientCh.output <- ""
					err := errors.New("ERROR: Not enough items\n")
					clientCh.err <- err
				}()
				continue
			}
			key := cmd_list[1]
			value := cmd_list[2]
			data[key] = value
			go func() {
				clientCh.output <- value
				if disk {
					str := "SET " + key + " " + value + "\n"
					saveCh.saveData <- str
					clientCh.err <- <-saveCh.err
				} else {
					clientCh.err <- nil
				}
			}()
		// GET <KEY>
		case "GET":
			key := cmd_list[1]
			value, ok := data[key]
			if ok {
				go func() {
					clientCh.output <- value
					clientCh.err <- nil
				}()
			} else {
				go func() {
					clientCh.output <- ""
					err := errors.New("ERROR: Unknown key \"" + key + "\"\n")
					clientCh.err <- err
				}()
			}
		// DEL <KEY>
		case "DEL":
			key := cmd_list[1]
			value, ok := data[key]
			if ok {
				delete(data, key)
				go func() {
					clientCh.output <- value
					if disk {
						str := "DEL " + key + "\n"
						saveCh.saveData <- str
						clientCh.err <- <-saveCh.err
					} else {
						clientCh.err <- nil
					}
				}()
			} else {
				go func() {
					clientCh.output <- ""
					err := errors.New("ERROR: Unknown key \"" + key + "\"\n")
					clientCh.err <- err
				}()
			}
		default:
			// sending error
			go func() {
				clientCh.output <- ""
				err := errors.New("ERROR: Unknown command\n")
				clientCh.err <- err
			}()
		}

	}
}

func saveData(file_pass string, saveCh saveChan) {
	file, err := os.Create(file_pass)
	defer file.Close()
	if err != nil {
		saveCh.err <- err
	} else {
		saveCh.err <- nil
		for str := range saveCh.saveData {
			_, err := file.WriteString(str)
			saveCh.err <- err
		}
	}
}

func handle(conn net.Conn, clientCh clientChan) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		command := scanner.Text()
		clientCh.input <- command

		answer := <-clientCh.output
		err := <-clientCh.err
		if err != nil {
			io.WriteString(conn, err.Error())
		} else {
			answer += "\n"
			io.WriteString(conn, answer)
		}
	}
}

type clientChan struct {
	input  chan string
	output chan string
	err    chan error
}

type saveChan struct {
	saveData chan string
	err      chan error
}

func main() {
	port, disk, file_pass, cmd_err := startup()


	//No errors, open port, start redis, start listening
	if cmd_err == nil {
		li, net_err := net.Listen("tcp", port)
		if net_err != nil {
			log.Fatalln(net_err)
		} else {
			defer li.Close()

			//Prepare channels for client and saving data
			cl_input := make(chan string)
			cl_output := make(chan string)
			cl_err := make(chan error)
			sv_data := make(chan string)
			sv_err := make(chan error)
			var clientCh clientChan
			clientCh.input = cl_input
			clientCh.output = cl_output
			clientCh.err = cl_err
			var saveCh saveChan
			saveCh.saveData = sv_data
			saveCh.err = sv_err

			if disk {
				go saveData(file_pass, saveCh)
				err := <-saveCh.err
				if err == nil {
					go redis(clientCh, saveCh, disk)
				}
			} else {
				go redis(clientCh, saveCh, disk)
			}

			for {
				conn, err := li.Accept()
				if err != nil {
					log.Fatalln(err)
				}

				go handle(conn, clientCh)
			}
		}
	} else {
	//Some CMD error
		fmt.Println(cmd_err)
	}
}