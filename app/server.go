package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		connection, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(connection)
	}

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

func handleConnection(connection net.Conn) {
	defer connection.Close()
	requestStruct, err := ParseRequestFromConnection(connection)
	if err != nil {
		fmt.Println("error in parsing request from connection", err.Error())
		os.Exit(1)
	}

	if strings.Contains(requestStruct.Path, "/files") {
		stage7Resp := getStage7Response(requestStruct)
		resp, err := CreateResponseFromResponseStruct(stage7Resp)
		if err != nil {
			fmt.Println("error in creating response from response struct", err.Error())
			os.Exit(1)
		}
		connection.Write([]byte(resp))
	}

	// below we send stage 5 response
	if strings.Contains(requestStruct.Path, "/user-agent") {
		stage5Resp := getStage5Response(requestStruct)
		resp, err := CreateResponseFromResponseStruct(stage5Resp)
		if err != nil {
			fmt.Println("error in creating response from response struct", err.Error())
			os.Exit(1)
		}
		connection.Write([]byte(resp))
		return
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
		Headers:     make(map[string]string),
	}
	// parsing the headers
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("error in reading header line from reader")
		}
		line = strings.TrimSuffix(line, "\r\n")
		// an empty line indicates the end of the headers
		if line == "" {
			break
		}

		// split the header into name and value
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("error in parsing header line: %s", line)
		}
		requestStruct.Headers[parts[0]] = parts[1]
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

func getStage5Response(reqStruct *RequestStruct) *ResponseStruct {
	userAgentHeaderData := reqStruct.Headers["User-Agent"]
	return &ResponseStruct{
		HttpVersion:       "HTTP/1.1",
		HttpStatusCode:    "200",
		HttpStatusMessage: "OK",
		Headers: map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(userAgentHeaderData)),
		},
		Body: []byte(userAgentHeaderData),
	}
}

func getStage7Response(reqStruct *RequestStruct) *ResponseStruct {
	directory := os.Args[2]
	fileName := strings.Split(reqStruct.Path, "/")[2]
	fullPath := path.Join(directory, fileName)
	fileData, err := os.ReadFile(fullPath)
	if err != nil {
		return &ResponseStruct{
			HttpVersion:       "HTTP/1.1",
			HttpStatusCode:    "404",
			HttpStatusMessage: "Not Found",
		}
	}
	return &ResponseStruct{
		HttpVersion:       "HTTP/1.1",
		HttpStatusCode:    "200",
		HttpStatusMessage: "OK",
		Headers: map[string]string{
			"Content-Type":   "application/octet-stream",
			"Content-Length": fmt.Sprintf("%d", len(fileData)),
		},
		Body: []byte(fileData),
	}
}
