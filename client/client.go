package client

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
)

type rtspConn struct {
	c               *net.TCPConn
	domain          string
	port            int
	username        string
	password        string
	domainport      string
	uri             string
	nonce           string
	realm           string
	command         string
	expectedEndLine string
}

func newRtspConn(domain string, port int, username, password string) rtspConn {
	domainport := fmt.Sprintf("%s:%d", domain, port)
	return rtspConn{
		c:               connectionOpen(domainport),
		domain:          domain,
		port:            port,
		username:        username,
		password:        password,
		uri:             fmt.Sprintf("rtsp://%s/videoMain", domainport),
		domainport:      domainport,
		nonce:           "",
		realm:           "",
		command:         "",
		expectedEndLine: ""}
}

//Client - basic tcp client to make rtsp calls to a home foscam
func Client(domain string, port int, username, password string) {
	r := newRtspConn(domain, port, username, password)
	for {
		fmt.Printf("main_loop\n")
		input := getUserInput()
		r = getCommand(r, input)
		if r.command == "end" {
			break
		} else {
			connectionWrite(r)
			r = connectionRead(r)
		}
	}
	fmt.Printf("end main_loop\n")
	connectionClose(r)
}

func connectionOpen(domainport string) *net.TCPConn {
	fmt.Printf("CLIENT: Opening connection to %s...\n", domainport)

	addr, err := net.ResolveTCPAddr("tcp", domainport)
	if err != nil {
		fmt.Printf("ResolveTCPAddr Error: %s\n", err.Error())
	}

	c, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		fmt.Printf("DialTCP Error: %s\n", err.Error())
	}
	fmt.Printf("Connected to %s\n", domainport)
	return c
}

func connectionClose(r rtspConn) {
	fmt.Printf("...Closing connection to %s\n", r.domainport)
	r.c.Close()
}

func connectionWrite(r rtspConn) {
	count, err := r.c.Write([]byte(r.command))
	if err != nil {
		fmt.Println("Write Error: ", err, count)
	}
	fmt.Printf("Sent command %d chars\n%s\n", count, r.command)
}

func connectionRead(r rtspConn) rtspConn {
	scanner := bufio.NewScanner(bufio.NewReader(r.c))
	data := ""
	line := ""
	for scanner.Scan() {
		line = scanner.Text()
		data += line + "\n"
		fmt.Printf("LINE:%s\n", line)
		if line == r.expectedEndLine {
			break
		}
	}
	fmt.Printf("RECEIVED:\n%s", data)

	//set properties if the server response contains them
	r.nonce = getNamedQuotedValue(r.nonce, data, "nonce")
	r.realm = getNamedQuotedValue(r.realm, data, "realm")
	return r
}

//getNamedQuotedValue gets the string from within quotes
//by finding the name in a supplied string e.g. if name is "realm"
//realm = "Testy" would return "Testy"
func getNamedQuotedValue(defaultValue, data, name string) string {
	val := defaultValue

	word := strings.Index(data, name)
	if word != -1 {
		firstQuote := strings.Index(data[word:], "\"")

		if firstQuote != -1 {
			secondQuote := strings.Index(data[word+firstQuote+1:], "\"")

			if secondQuote != -1 {
				val = data[word+firstQuote+1 : word+firstQuote+1+secondQuote]
			}
		}
	}

	return val
}

func getUserInput() int {
	input := 0
	fmt.Println("1:OPTIONS 2:DESCRIBE 3:AUTH 9:end")
	fmt.Scan(&input)
	return input
}

func getCommand(r rtspConn, input int) rtspConn {
	switch input {
	case 1:
		r.command = fmt.Sprintf("OPTIONS %s RTSP/1.0\r\nCSeq: 1\r\n\r\n", r.uri)
		r.expectedEndLine = ""
		/*
			Client Request
			>>>>>>>>>>>>>>>>>>>>>>>
			OPTIONS rtsp://127.0.0.1:5000/videoMain RTSP/1.0
			CSeq: 1

			Server Response example
			>>>>>>>>>>>>>>>>>>>>>>>
			RTSP/1.0 200 OK
			CSeq: 1
			Date: Sun, Mar 29 2020 14:50:01 GMT
			Public: OPTIONS, DESCRIBE, SETUP, TEARDOWN, PLAY, PAUSE, GET_PARAMETER, SET_PARAMETER
		*/
	case 2:
		r.command = fmt.Sprintf("DESCRIBE %s RTSP/1.0\r\nAccept: application/sdp\r\nCSeq: 2\r\n\r\n", r.uri)
		r.expectedEndLine = ""
		/*
			Client Request
			>>>>>>>>>>>>>>>>>>>>>>>
			DESCRIBE rtsp://192.168.1.11:88/videoMain RTSP/1.0
			Accept: application/sdp
			CSeq: 2

			Server Response example
			>>>>>>>>>>>>>>>>>>>>>>>
			RTSP/1.0 401 Unauthorized
			CSeq: 2
			Date: Sun, Mar 29 2020 14:50:02 GMT
			WWW-Authenticate: Digest realm="Foscam IPCam Living Video", nonce="f77741eb931f26a0aaf3a01a0de5944f"
		*/
	case 3:
		method := "DESCRIBE"
		response := getResponse(r, method)
		r.command = fmt.Sprintf("%s %s RTSP/1.0\r\nAccept: application/sdp\r\nCSeq: 3\r\nAuthorization: Digest username=\"%s\", realm=\"%s\", nonce=\"%s\", uri=\"%s\", response=\"%s\"\r\n\r\n", method, r.uri, r.username, r.realm, r.nonce, r.uri, response)
		r.expectedEndLine = "a=control:track2"
		/*
			Client Request
			>>>>>>>>>>>>>>>>>>>>>>>
			DESCRIBE rtsp://192.168.1.11:88/videoMain RTSP/1.0
			Accept: application/sdp
			CSeq: 3
			Authorization: Digest username="<your_username>", realm="Foscam IPCam Living Video", nonce="f77741eb931f26a0aaf3a01a0de5944f", uri="rtsp://192.168.1.11:88/videoMain", response="aaaabbbbccccddddeeeeffffgggghhhh"

			Server Response example
			>>>>>>>>>>>>>>>>>>>>>>>
			RTSP/1.0 200 OK
			CSeq: 3
			Date: Sun, Mar 29 2020 04:55:49 GMT
			Content-Base: rtsp://192.168.1.11:65534/videoMain/
			Content-Type: application/sdp
			Content-Length: 543

			v=0
			o=- 1582535758913729 1 IN IP4 192.168.1.11
			s=IP Camera Video
			i=videoMain
			t=0 0
			a=tool:LIVE555 Streaming Media v2014.02.10
			a=type:broadcast
			a=control:*
			a=range:npt=0-
			a=x-qt-text-nam:IP Camera Video
			a=x-qt-text-inf:videoMain
			m=video 0 RTP/AVP 96
			c=IN IP4 0.0.0.0
			b=AS:96
			a=rtpmap:96 H264/90000
			a=fmtp:96 packetization-mode=1;profile-level-id=64001F;sprop-parameter-sets=Z2QAH6w0zAUAW/8BagICAoAAAfRo7jwwdDACgoACgiXeXGhgBQUABQRLvLhQ,aO48MA==
			a=control:track1
			m=audio 0 RTP/AVP 0
			c=IN IP4 0.0.0.0
			b=AS:64
			a=control:track2
		*/
	case 9:
		r.command = "end"
		r.expectedEndLine = ""
	}
	return r
}

func getResponse(r rtspConn, method string) string {
	//https://stackoverflow.com/questions/55379440/rtsp-video-streaming-with-authentication
	//https://mrwaggel.be/post/golang-hash-sum-and-checksum-to-string-tutorial-and-examples/
	ha1 := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", r.username, r.realm, r.password)))
	ha2 := md5.Sum([]byte(fmt.Sprintf("%s:rtsp://%s/videoMain", method, r.domainport)))
	ha3 := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", hex.EncodeToString(ha1[:]), r.nonce, hex.EncodeToString(ha2[:]))))
	return hex.EncodeToString(ha3[:])
}
