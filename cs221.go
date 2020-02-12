package cs221

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
    "time"
)

type Conn struct {
	In <-chan string
	Out chan<- string
}

func Lines(message string) []string {
	ls := strings.Split(message,"\n")
	return ls[:len(ls)-2]
}

func HeadLine(message string) string {
	return Lines(message)[0]
}

func takeDictation(comm <-chan string, who string) string {
	message := ""
	sendbuffer := true
	for sendbuffer {
		next := <-comm
		if next == "" {
			return ""
		}
		message += next
		if message == "\n" || strings.HasSuffix(message, "\n\n") {
			where := strings.Index(message, "\n\n")
			if where != -1 && where < len(message)-2 {
				fmt.Println(who+" sent several empty lines.")
				panic(who+" sent several empty lines.")
			}
			sendbuffer = false
		}
	}

	return message
}

func listenOnWire(conn net.Conn, who string) string {
	reader := bufio.NewReader(conn)
	reply := ""
	replybuffer := true
	for replybuffer {
		next, err := reader.ReadString('\n')
		if err != nil {
			return ""
		}
		reply += next
		if next == "\n" {
			replybuffer = false
		}
	}
	return reply
}

func makeServerProxy(conn net.Conn, who string) (chan<- string, <-chan string) {
	server := make(chan string)
	client := make(chan string)
	go func() {
		done := false
		for !done {
			// Get a client's request from the network.
			request := listenOnWire(conn, who)
			// Forward it over the local channel.
			client <- request
			if request !=  "" {
				// Gather the server's response from the channel.
				response := takeDictation(server, who)
				// Forward it to the client over the network.
				fmt.Fprintf(conn, response)
			} else {
				done = true
			}
		}
	}()
	return server, client
}

func makeClientProxy(conn net.Conn, who string) (chan<- string, <-chan string) {
	server := make(chan string)
	client := make(chan string)
	go func() {
		done := false
		for !done {
			// Gather a client's request from the local channel.
			request := takeDictation(client,who)
			if request !=  "" {
				// Forward it to the server over the network.
				fmt.Fprintf(conn, request)

				// Gather the server's response from the network.
				response := listenOnWire(conn, who)
				if response != "" {
					// Forward it to the client over the local channel.
					server <- response
				} else {
					done = true
				}
			} else {
				done  = true
			}
		}
	}()
	return client, server
}

func MakeConnection(hostname string, port int, who string) (chan<- string, <-chan string, error) {
	connection := hostname + ":" + strconv.Itoa(port)
	conn, err := net.DialTimeout("tcp", connection, time.Second)
	if err != nil {
		return nil, nil, err
	}
	outgoing,incoming := makeClientProxy(conn, who)
	return outgoing, incoming, nil
}

func HandleConnections(hostname string, port int, 
	handler func(chan<- string,<-chan string,interface{}), 
	who string, sharedinfo interface{}) error {

	connection := hostname + ":" + strconv.Itoa(port)
	// Listen for incoming connections.
	lis, err := net.Listen("tcp", connection)

	if err != nil {
		return err
	}

	// Close the listener when the application closes.
	defer lis.Close()

	for {
		// Accept an incoming connection.
		conn, err := lis.Accept()
		if err != nil {
			return err
		}

		outgoing,incoming := makeServerProxy(conn, who)
		go handler(outgoing,incoming,sharedinfo)
	}
}


func HandleAllConnections(hostname string, port int) (chan Conn,error) {

	connection := hostname + ":" + strconv.Itoa(port)
	// Listen for incoming connections.
	lis, err := net.Listen("tcp", connection)

	if err != nil {
		return nil,err
	}

	// Close the listener when the application closes.

	c := make(chan Conn)
	go func() {
		defer lis.Close()
		for {
			// Accept an incoming connection.
			conn, err := lis.Accept()
			if err != nil {
				continue
			} else {
				outgoing,incoming := makeServerProxy(conn, "proxy")
				c <- Conn{Out:outgoing,In:incoming}
			}
		}
	} ()

	return c,nil
}

