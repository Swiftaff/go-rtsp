package client

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type rtspConn struct {
	c               *net.TCPConn
	localDomain     string
	localPort1      int
	localPort2      int
	localPortRange  string
	domain          string
	port            int
	remotePort      string
	remoteURI       string
	remoteURItrack1 string
	remoteURItrack2 string
	username        string
	password        string
	domainport      string
	uri             string
	nonce           string
	realm           string
	command         string
	expectedEndLine string
	session         string
}

func getLocalDomainPort(c *net.TCPConn) (string, int) {
	addr := c.LocalAddr()
	addrAsString := addr.String()
	addrAsSlice := strings.Split(addrAsString, ":")
	localDomain := strings.Join([]string{addrAsSlice[0]}, "")
	localPort1AsString := strings.Join([]string{addrAsSlice[1]}, "")
	localPort1, _ := strconv.Atoi(localPort1AsString)
	return localDomain, localPort1
}

func newRtspConn(domain string, port int, username, password string) rtspConn {
	domainport := fmt.Sprintf("%s:%d", domain, port)
	c := connectionOpen(domainport)
	localDomain, localPort1 := getLocalDomainPort(c)
	localPort2 := localPort1 + 1
	localPortRange := fmt.Sprintf("%d-%d", localPort1, localPort2)
	return rtspConn{
		c:               c,
		localDomain:     localDomain,
		localPort1:      localPort1,
		localPort2:      localPort2,
		localPortRange:  localPortRange,
		domain:          domain,
		port:            port,
		remotePort:      "",
		remoteURI:       "",
		remoteURItrack1: "",
		remoteURItrack2: "",
		username:        username,
		password:        password,
		uri:             fmt.Sprintf("rtsp://%s/videoMain", domainport),
		domainport:      domainport,
		nonce:           "",
		realm:           "",
		command:         "",
		expectedEndLine: "",
		session:         ""}
}

//Client - basic tcp client to make rtsp calls to a home foscam
//calls main handshake methods sequentially, automatically
func Client(domain string, port int, username, password string) {
	r := newRtspConn(domain, port, username, password)
	port = r.c.RemoteAddr().(*net.TCPAddr).Port
	commands := []int{1, 2, 3, 4, 6}
	for i := 0; i < len(commands); i++ {
		r = getCommand(r, commands[i])
		connectionWrite(r)
		r = connectionRead(r)
	}
	//connectionClose(r)
}

//ManualClient - basic tcp client to make rtsp calls to a home foscam
//you need to type the key for each method, manually
func ManualClient(domain string, port int, username, password string) {
	r := newRtspConn(domain, port, username, password)
	port = r.c.RemoteAddr().(*net.TCPAddr).Port
	for {
		input := getUserInput()
		r = getCommand(r, input)
		if r.command == "end" {
			break
		} else {
			connectionWrite(r)
			r = connectionRead(r)
		}
	}
	connectionClose(r)
}

func connectionOpen(domainport string) *net.TCPConn {
	fmt.Printf("CLIENT: Opening connection to %s...\n", domainport)

	server, err := net.ResolveTCPAddr("tcp", domainport)
	if err != nil {
		fmt.Printf("ResolveTCPAddr Error: %s\n", err.Error())
	}
	c, err := net.DialTCP("tcp", nil, server)
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
		if line == r.expectedEndLine {
			break
		}
	}
	fmt.Printf("RECEIVED:\n%s", data)

	//set properties if the server response contains them
	r.nonce = getNamedQuotedValue(r.nonce, data, "nonce")
	r.realm = getNamedQuotedValue(r.realm, data, "realm")
	r.session = getNamedColonValue(r.session, data, "Session")

	url := getNamedColonValue("", data, "Content-Base")
	r.remotePort = getPortFromURL("", url)
	r.remoteURI = fmt.Sprintf("rtsp://%s:%s/videoMain/", r.domain, r.remotePort)
	r.remoteURItrack1 = fmt.Sprintf("rtsp://%s:%s/videoMain/track1", r.domain, r.remotePort)
	r.remoteURItrack2 = fmt.Sprintf("rtsp://%s:%s/videoMain/track2", r.domain, r.remotePort)
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

//getNamedColonValue gets the string following the colon
//after finding the name in a supplied string e.g. if name is "realm"
//realm: Testy\n would return "Testy"
func getNamedColonValue(defaultValue, data, name string) string {
	val := defaultValue
	word := strings.Index(data, name)
	if word != -1 {
		start := word + len(name) + 2
		length := strings.Index(data[start:], "\n")
		if length != -1 {
			val = data[start : start+length]
		}
	}
	return val
}

//getPortFromURL gets the string following the second colon in a url
//rtsp://testy:1234/testy2\n would return string "1234"
func getPortFromURL(defaultValue, data string) string {
	val := defaultValue
	firstColon := strings.Index(data, ":")
	if firstColon != -1 {
		secondColon := strings.Index(data[firstColon+1:], ":")
		nextForwardSlash := strings.Index(data[firstColon+secondColon+1:], "/")
		endOfLine := strings.Index(data[firstColon+secondColon+1:], "\n")
		if nextForwardSlash != -1 {
			endOfLine = nextForwardSlash
		}
		if endOfLine == -1 {
			val = data[firstColon+secondColon:]
		} else {
			val = data[firstColon+secondColon : firstColon+secondColon+endOfLine]
		}

	}
	return val
}

func getUserInput() int {
	input := 0
	fmt.Println("1:OPTIONS 2:DESCRIBE 3:AUTH 4:SETUP(video) 5:SETUP(audio not done yet) 6:PLAY 9:end")
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
		response := getResponse(r, method, false)
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
	case 4:
		method := "SETUP"
		response := getResponse(r, method, true)
		r.command = fmt.Sprintf("%s %s RTSP/1.0\r\nTransport: RTP/AVP/UDP;unicast;client_port=%s\r\nCSeq: 4\r\nAuthorization: Digest username=\"%s\", realm=\"%s\", nonce=\"%s\", uri=\"%s\", response=\"%s\"\r\n\r\n", method, r.remoteURItrack1, r.localPortRange, r.username, r.realm, r.nonce, r.remoteURItrack1, response)
		r.expectedEndLine = ""
		/*
			Client Request
			>>>>>>>>>>>>>>>>>>>>>>>
			SETUP rtsp://192.168.1.11:65534/videoMain/track1 RTSP/1.0
			Transport: RTP/AVP/UDP;unicast;client_port=22292-22293
			CSeq: 4
			Authorization: Digest username="foscam", realm="Foscam IPCam Living Video", nonce="d4dea19512f55715153c33444c826135", uri="rtsp://192.168.1.11:65534/videoMain/track1", response="041d36dbb3f0b82a18d8335775aad078"

			Server Response example
			>>>>>>>>>>>>>>>>>>>>>>>
			RTSP/1.0 200 OK
			CSeq: 4
			Date: Sun, Mar 29 2020 04:55:49 GMT
			Transport: RTP/AVP;unicast;destination=192.168.1.157;source=192.168.1.11;client_port=22292-22293;server_port=6970-6971
			Session: 59F3B34B;timeout=65
		*/

	//skip track2 for now...
	/*
		case 5:
			method := "SETUP"
			response := getResponse(r, method, true)
			r.command = fmt.Sprintf("%s %s RTSP/1.0\r\nTransport: RTP/AVP/UDP;unicast;client_port=%s\r\nCSeq: 4\r\nAuthorization: Digest username=\"%s\", realm=\"%s\", nonce=\"%s\", uri=\"%s\", response=\"%s\"\r\n\r\n", method, r.remoteURI, r.localPortRange, r.username, r.realm, r.nonce, r.remoteURI, response)
			r.expectedEndLine = ""
			/*
			SETUP rtsp://192.168.1.11:65534/videoMain/track2 RTSP/1.0
			Transport: RTP/AVP/UDP;unicast;client_port=22294-22295
			CSeq: 5
			Session: 59F3B34B
			Authorization: Digest username="foscam", realm="Foscam IPCam Living Video", nonce="d4dea19512f55715153c33444c826135", uri="rtsp://192.168.1.11:65534/videoMain/track2", response="43549f4c211b5f71637cc0284c6647f3"


	*/

	case 6:
		method := "PLAY"
		response := getResponse(r, method, true)
		r.command = fmt.Sprintf("%s %s RTSP/1.0\r\nRange: npt=0-\r\nCSeq: 6\r\nSession: %s\r\nAuthorization: Digest username=\"%s\", realm=\"%s\", nonce=\"%s\", uri=\"%s\", response=\"%s\"\r\n\r\n", method, r.remoteURI, r.session, r.username, r.realm, r.nonce, r.remoteURI, response)
		r.expectedEndLine = ""
		/*
			Client Request
			>>>>>>>>>>>>>>>>>>>>>>>
			first...
			RDT
			RTCP
			UDP
			RTCP

			then...
			PLAY rtsp://192.168.1.11:65534/videoMain/ RTSP/1.0
			Range: npt=0.000-
			CSeq: 6
			Session: 59F3B34B
			Authorization: Digest username="foscam", realm="Foscam IPCam Living Video", nonce="d4dea19512f55715153c33444c826135", uri="rtsp://192.168.1.11:65534/videoMain/", response="f12c2411ddb14be3eb21fd9d0c3528c9"


			Server Response example
			>>>>>>>>>>>>>>>>>>>>>>>
			RTSP/1.0 200 OK
			CSeq: 6
			Date: Sun, Mar 29 2020 04:55:49 GMT
			Range: npt=0.000-
			Session: 59F3B34B
			RTP-Info: url=rtsp://192.168.1.11:65534/videoMain/track1;seq=14200;rtptime=1889286248,url=rtsp://192.168.1.11:65534/videoMain/track2;seq=44635;rtptime=2157729542


		*/
	case 9:
		r.command = "end"
		r.expectedEndLine = ""
	}
	return r
}

func getResponse(r rtspConn, method string, useSecondDomain bool) string {
	//https://stackoverflow.com/questions/55379440/rtsp-video-streaming-with-authentication
	//https://mrwaggel.be/post/golang-hash-sum-and-checksum-to-string-tutorial-and-examples/
	ha1 := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", r.username, r.realm, r.password)))
	ha2 := md5.Sum([]byte(fmt.Sprintf("%s:rtsp://%s/videoMain", method, r.domainport)))
	if useSecondDomain {
		ha2 = md5.Sum([]byte(fmt.Sprintf("%s:%s", method, r.remoteURI)))
	}
	ha3 := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", hex.EncodeToString(ha1[:]), r.nonce, hex.EncodeToString(ha2[:]))))
	return hex.EncodeToString(ha3[:])
}
