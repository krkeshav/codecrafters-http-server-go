package main

import (
	"bufio"
	"bytes"
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

	// below we send response according to stage 4
	if strings.Contains(requestStruct.Path, "/echo/") {
		stage4Resp := getStage4Response(requestStruct)
		resp, err := CreateResponseFromResponseStruct(stage4Resp)
		if err != nil {
			fmt.Println("error in creating response from response struct", err.Error())
			os.Exit(1)
		}
		connection.Write([]byte(resp))
		return
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

func CreateResponseFromResponseStruct(respStruct *ResponseStruct) ([]byte, error) {
	writer := bytes.NewBuffer(make([]byte, 0))
	writer.WriteString(respStruct.HttpVersion)
	writer.WriteString(" ")
	writer.WriteString(respStruct.HttpStatusCode)
	writer.WriteString(" ")
	writer.WriteString(respStruct.HttpStatusMessage)
	writer.WriteString("\r\n")

	for headerKey, headerValue := range respStruct.Headers {
		writer.WriteString(fmt.Sprintf("%s: %s", headerKey, headerValue))
		writer.WriteString("\r\n")
	}

	writer.WriteString("\r\n")
	if len(respStruct.Body) > 0 {
		writer.WriteString(string(respStruct.Body))
	}

	return writer.Bytes(), nil
}

func getStage4Response(reqStruct *RequestStruct) *ResponseStruct {
	splitPath := strings.SplitN(reqStruct.Path, "/echo/", 2)
	fmt.Println(splitPath)
	respBodyString := splitPath[1]

	return &ResponseStruct{
		HttpVersion:       "HTTP/1.1",
		HttpStatusCode:    "200",
		HttpStatusMessage: "OK",
		Headers: map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(respBodyString)),
		},
		Body: []byte(respBodyString),
	}
}
