package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	connection, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	// close the connection after everything is handled
	defer connection.Close()

	requestStruct, err := ParseRequestFromConnection(connection)
	if err != nil {
		fmt.Println("error in parsing request from connection", err.Error())
		os.Exit(1)
	}

	// below check is only for stage 3
	if requestStruct.Path != "/" {
		httpNotFoundMessage := "HTTP/1.1 404 Not Found\r\n\r\n"
		connection.Write([]byte(httpNotFoundMessage))
		return
	}
	// dummy http response sent through connection
	httpOkMessage := "HTTP/1.1 200 OK\r\n\r\n"
	connection.Write([]byte(httpOkMessage))
}

type RequestStruct struct {
	Method      string
	Path        string
	HttpVersion string
	Headers     map[string]string
	Body        []byte
}

type ResponseStruct struct {
	HttpVersion       string
	HttpStatusCode    string
	HttpStatusMessage string
	Headers           map[string]string
	Body              []byte
}

func ParseRequestFromConnection(connection net.Conn) (*RequestStruct, error) {
	reader := bufio.NewReader(connection)
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error in reading status line from reader")
	}
	statusLine = strings.TrimSuffix(statusLine, "\r\n")
	statusLineComponents := strings.Split(statusLine, " ")
	if len(statusLineComponents) < 3 {
		return nil, fmt.Errorf("error in parsing status line components from status line")
	}
	requestStruct := &RequestStruct{
		Method:      statusLineComponents[0],
		Path:        statusLineComponents[1],
		HttpVersion: statusLineComponents[2],
	}

	return requestStruct, nil
}
