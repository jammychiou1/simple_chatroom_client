package main

import (
	"bufio"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

type RegisterInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Result struct {
	Result interface{} `json:"result"`
}

type Server struct {
	conn net.Conn
	r    *bufio.Reader
	w    *bufio.Writer
}

type ChatroomInfo struct {
	ID     int    `json:"id"`
	Friend string `json:"friend"`
}

type Message struct {
	From  string `json:"from"`
	Type  string `json:"type"`
	Data  string `json:"data"`
	Token string `json:"token"`
}

func connect() (Server, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", os.Args[1], os.Args[2]))
	if err != nil {
		return Server{}, err
	}
	return Server{conn: conn, r: bufio.NewReader(conn), w: bufio.NewWriter(conn)}, nil
}

func handleRegister(clnt webClient) {
	var regInfo RegisterInfo
	if err := json.Unmarshal(clnt.body, &regInfo); err != nil {
		//log.Println(err)
		writeHeader(clnt, 400, 0)
		return
	}
	srvr, err := connect()
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	defer srvr.conn.Close()
	fmt.Fprintf(srvr.w, "register %s %s\n", b64.StdEncoding.EncodeToString([]byte(regInfo.Username)), b64.StdEncoding.EncodeToString([]byte(regInfo.Password)))
	srvr.w.Flush()
	ln, _ := srvr.r.ReadString('\n')
	rslt := Result{Result: strings.TrimSpace(ln)}
	rspBytes, _ := json.Marshal(rslt)
	writeHeader(clnt, 200, len(rspBytes))
	clnt.w.Write(rspBytes)
	clnt.w.Flush()
}
func handleCheck(clnt webClient) {
	srvr, err := connect()
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	defer srvr.conn.Close()
	fmt.Fprintf(srvr.w, "login %s %s\n", b64.StdEncoding.EncodeToString([]byte(clnt.username)), b64.StdEncoding.EncodeToString([]byte(clnt.password)))
	srvr.w.Flush()
	ln, _ := srvr.r.ReadString('\n')
	rslt := Result{Result: strings.TrimSpace(ln)}
	rspBytes, _ := json.Marshal(rslt)
	writeHeader(clnt, 200, len(rspBytes))
	clnt.w.Write(rspBytes)
	clnt.w.Flush()
}
func handleListChatrooms(clnt webClient) {
	srvr, err := connect()
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	defer srvr.conn.Close()
	fmt.Fprintf(srvr.w, "login %s %s\n", b64.StdEncoding.EncodeToString([]byte(clnt.username)), b64.StdEncoding.EncodeToString([]byte(clnt.password)))
	srvr.w.Flush()
	srvr.r.ReadString('\n')
	fmt.Fprintf(srvr.w, "listChatroom\n")
	srvr.w.Flush()

	res, err := srvr.r.ReadString('\n')
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	res = strings.Replace(res, "\n", "", -1)
	//log.Println(res)

	chatroomNum, err := strconv.Atoi(res)
	chatroomInfos := make([]ChatroomInfo, chatroomNum, chatroomNum)
	for i := 0; i < chatroomNum; i++ {
		res, err := srvr.r.ReadString('\n')
		if err != nil {
			writeHeader(clnt, 400, 0)
			return
		}
		res = strings.Replace(res, "\n", "", -1)
		chatroomInfo := strings.Split(res, " ")
		id, err := strconv.Atoi(chatroomInfo[0])
		if err != nil {
			writeHeader(clnt, 400, 0)
			return
		}
		mem1, err := b64.StdEncoding.DecodeString(chatroomInfo[1])
		if err != nil {
			writeHeader(clnt, 400, 0)
			return
		}
		mem2, err := b64.StdEncoding.DecodeString(chatroomInfo[2])
		if err != nil {
			writeHeader(clnt, 400, 0)
			return
		}
		fmt.Printf("(" + chatroomInfo[0] + ") ")
		var friend string
		if clnt.username == string(mem1) {
			friend = string(mem2)
		} else {
			friend = string(mem1)
		}
		chatroomInfos[i] = ChatroomInfo{ID: id, Friend: friend}
	}

	rslt := Result{Result: chatroomInfos}
	rspBytes, _ := json.Marshal(rslt)
	writeHeader(clnt, 200, len(rspBytes))
	clnt.w.Write(rspBytes)
	clnt.w.Flush()
}
func handleAddChatrooms(clnt webClient) {
	srvr, err := connect()
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	defer srvr.conn.Close()
	fmt.Fprintf(srvr.w, "login %s %s\n", b64.StdEncoding.EncodeToString([]byte(clnt.username)), b64.StdEncoding.EncodeToString([]byte(clnt.password)))
	srvr.w.Flush()
	srvr.r.ReadString('\n')
	var obj map[string]string
	if err := json.Unmarshal(clnt.body, &obj); err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	friend := obj["friend"]
	fmt.Fprintf(srvr.w, "createChatroom %s\n", b64.StdEncoding.EncodeToString([]byte(friend)))
	srvr.w.Flush()
	res, err := srvr.r.ReadString('\n')
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	res = strings.TrimSpace(res)
	tokens := strings.Split(res, " ")
	if len(tokens) == 0 {
		writeHeader(clnt, 400, 0)
		return
	}
	if tokens[0] == "ok" {
		id, _ := strconv.Atoi(tokens[1])
		rslt := Result{Result: id}
		rspBytes, _ := json.Marshal(rslt)
		writeHeader(clnt, 200, len(rspBytes))
		clnt.w.Write(rspBytes)
		clnt.w.Flush()
		return
	} else {
		rslt := Result{Result: -1}
		rspBytes, _ := json.Marshal(rslt)
		writeHeader(clnt, 200, len(rspBytes))
		clnt.w.Write(rspBytes)
		clnt.w.Flush()
		return
	}
}
func handleListFriends(clnt webClient) {
	srvr, err := connect()
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	defer srvr.conn.Close()
	fmt.Fprintf(srvr.w, "login %s %s\n", b64.StdEncoding.EncodeToString([]byte(clnt.username)), b64.StdEncoding.EncodeToString([]byte(clnt.password)))
	srvr.w.Flush()
	srvr.r.ReadString('\n')
	fmt.Fprintf(srvr.w, "listFriends\n")
	srvr.w.Flush()

	res, err := srvr.r.ReadString('\n')
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	res = strings.Replace(res, "\n", "", -1)

	friendEncs := strings.Split(res, " ")
	friends := make([]string, len(friendEncs), len(friendEncs))
	for i, v := range friendEncs {
		friendBytes, err := b64.StdEncoding.DecodeString(v)
		if err != nil {
			writeHeader(clnt, 400, 0)
			return
		}
		friends[i] = string(friendBytes)
	}

	rslt := Result{Result: friends}
	rspBytes, _ := json.Marshal(rslt)
	writeHeader(clnt, 200, len(rspBytes))
	clnt.w.Write(rspBytes)
	clnt.w.Flush()
}
func handleAddFriend(clnt webClient) {
	srvr, err := connect()
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	defer srvr.conn.Close()
	fmt.Fprintf(srvr.w, "login %s %s\n", b64.StdEncoding.EncodeToString([]byte(clnt.username)), b64.StdEncoding.EncodeToString([]byte(clnt.password)))
	srvr.w.Flush()
	srvr.r.ReadString('\n')
	var obj map[string]string
	if err := json.Unmarshal(clnt.body, &obj); err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	friend := obj["friend"]
	fmt.Fprintf(srvr.w, "addFriend %s\n", b64.StdEncoding.EncodeToString([]byte(friend)))
	srvr.w.Flush()
	res, err := srvr.r.ReadString('\n')
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	res = strings.TrimSpace(res)

	rslt := Result{Result: res}
	rspBytes, _ := json.Marshal(rslt)
	writeHeader(clnt, 200, len(rspBytes))
	clnt.w.Write(rspBytes)
	clnt.w.Flush()
}
func handleDeleteFriend(clnt webClient) {
	srvr, err := connect()
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	defer srvr.conn.Close()
	fmt.Fprintf(srvr.w, "login %s %s\n", b64.StdEncoding.EncodeToString([]byte(clnt.username)), b64.StdEncoding.EncodeToString([]byte(clnt.password)))
	srvr.w.Flush()
	srvr.r.ReadString('\n')
	var obj map[string]string
	if err := json.Unmarshal(clnt.body, &obj); err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	friend := obj["friend"]
	fmt.Fprintf(srvr.w, "deleteFriend %s\n", b64.StdEncoding.EncodeToString([]byte(friend)))
	srvr.w.Flush()
	res, err := srvr.r.ReadString('\n')
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	res = strings.TrimSpace(res)

	rslt := Result{Result: res}
	rspBytes, _ := json.Marshal(rslt)
	writeHeader(clnt, 200, len(rspBytes))
	clnt.w.Write(rspBytes)
	clnt.w.Flush()
}

func handleListMessage(clnt webClient, srvr Server, chatroomID int) {
	//log.Println("list message")
	idx := strings.Index(clnt.path[3], "?begin=")
	if idx == -1 {
		writeHeader(clnt, 400, 0)
		return
	}
	idx += 7
	beg, err := strconv.Atoi(clnt.path[3][idx:])
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	fmt.Fprintf(srvr.w, "logs %d -1\n", beg)
	srvr.w.Flush()

	res, err := srvr.r.ReadString('\n')
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	res = strings.TrimSpace(res)
	//log.Printf("border %q", res)
	border := strings.Split(res, " ")
	left, err := strconv.Atoi(border[0])
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	right, err := strconv.Atoi(border[1])
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	//log.Printf("%d %d\n", left, right)
	messages := make([]Message, right-left, right-left)
	for i := left; i < right; i++ {
		res, err := srvr.r.ReadString('\n')
		errorHandler(err)
		res = strings.Replace(res, "\n", "", -1)
		tokens := strings.Split(res, " ")
		fromBytes, err := b64.StdEncoding.DecodeString(tokens[0])
		if err != nil {
			writeHeader(clnt, 400, 0)
			return
		}
		messages[i-left].From = string(fromBytes)
		messages[i-left].Type = tokens[1]
		dataBytes, err := b64.StdEncoding.DecodeString(tokens[2])
		if err != nil {
			writeHeader(clnt, 400, 0)
			return
		}
		messages[i-left].Data = string(dataBytes)
		if tokens[1] == "file" {
			messages[i-left].Token = tokens[3]
		} else if tokens[1] == "image" {
			messages[i-left].Token = tokens[3]
		}
	}

	rslt := Result{Result: messages}
	rspBytes, _ := json.Marshal(rslt)
	writeHeader(clnt, 200, len(rspBytes))
	clnt.w.Write(rspBytes)
	clnt.w.Flush()
}

func handleSendMessage(clnt webClient, srvr Server, chatroomID int) {
	var obj map[string]string
	if err := json.Unmarshal(clnt.body, &obj); err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	message := obj["message"]
	fmt.Fprintf(srvr.w, "sendMessage %s\n", b64.StdEncoding.EncodeToString([]byte(message)))
	srvr.w.Flush()

	rslt := Result{Result: "ok"}
	rspBytes, _ := json.Marshal(rslt)
	writeHeader(clnt, 200, len(rspBytes))
	clnt.w.Write(rspBytes)
	clnt.w.Flush()
}

func handleChatrooms(clnt webClient) {
	//log.Println("handle chatrooms")
	chatroomID, err := strconv.Atoi(clnt.path[2])
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}

	srvr, err := connect()
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	defer srvr.conn.Close()

	fmt.Fprintf(srvr.w, "login %s %s\n", b64.StdEncoding.EncodeToString([]byte(clnt.username)), b64.StdEncoding.EncodeToString([]byte(clnt.password)))
	srvr.w.Flush()
	srvr.r.ReadString('\n')
	fmt.Fprintf(srvr.w, "joinChatroom %d\n", chatroomID)
	srvr.w.Flush()
	res, err := srvr.r.ReadString('\n')
	//log.Printf("%q\n", res)
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	res = strings.TrimSpace(res)
	if res[:2] != "ok" {
		writeHeader(clnt, 400, 0)
		return
	}

	if clnt.method == "GET" {
		handleListMessage(clnt, srvr, chatroomID)
	} else {
		handleSendMessage(clnt, srvr, chatroomID)
	}
}
func handleStatic(clnt webClient) {
	fInfo, err := os.Stat("static/" + strings.Join(clnt.path[2:], "/"))
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	f, err := os.Open("static/" + strings.Join(clnt.path[2:], "/"))
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	writeHeader(clnt, 200, int(fInfo.Size()))
	io.CopyN(clnt.w, f, fInfo.Size())
	clnt.w.Flush()
}
func handleStreamFile(clnt webClient, isImage bool) {
	srvr, err := connect()
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	defer srvr.conn.Close()

	token := clnt.path[2]

	cmd := ""
	if isImage {
		cmd = "downloadImage"
	} else {
		cmd = "downloadFile"
	}
	fmt.Fprintf(srvr.w, "%s %s\n", cmd, token)
	srvr.w.Flush()
	ln, err := srvr.r.ReadString('\n')
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	ln = strings.TrimSpace(ln)
	tokens := strings.Split(ln, " ")
	if tokens[0] != "ok" {
		writeHeader(clnt, 400, 0)
		return
	}
	filenameBytes, err := b64.StdEncoding.DecodeString(tokens[1])
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	filename := string(filenameBytes)
	filesize, err := strconv.Atoi(tokens[2])
	if err != nil {
		writeHeader(clnt, 400, 0)
		return
	}
	if !isImage {
		clnt.rspHeaders["Content-Disposition"] = fmt.Sprintf("attachment; filename=\"%s\"", filename)
	}
	writeHeader(clnt, 200, filesize)
	io.CopyN(clnt.w, srvr.r, int64(filesize))
	clnt.w.Flush()
}
func handleDownloadFile(clnt webClient) {
	handleStreamFile(clnt, false)
}
func handleDownloadImage(clnt webClient) {
	handleStreamFile(clnt, true)
}
