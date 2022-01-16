package main

import (
	"bufio"
	b64 "encoding/base64"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

const BUFFER_SIZE = 1024

func main() {
	fmt.Println(os.Args[1])
	fmt.Println(fmt.Sprintf("%s:%s", os.Args[1], os.Args[2]))

	var wg sync.WaitGroup
    
    wg.Add(1)

    go func() {
        defer wg.Done()
        runWebserver()
    }()

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", os.Args[1], os.Args[2]))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var user string
	stdinReader := bufio.NewReader(os.Stdin)
	serverReader := bufio.NewReader(conn)
	// select login or register
	for {
		fmt.Println("Select options: \n(1) Login\n(2) Register")
		opt, err := stdinReader.ReadString('\n')
		errorHandler(err)
		opt = strings.Replace(opt, "\n", "", -1)

		if opt == "1" {
			// login
			// login: input username and password
			fmt.Print("login: \nusername: ")
			username, err := stdinReader.ReadString('\n')
			username = strings.Replace(username, "\n", "", -1)
			errorHandler(err)

			fmt.Print("password: ")
			password, err := stdinReader.ReadString('\n')
			password = strings.Replace(password, "\n", "", -1)
			errorHandler(err)

			usernameEnc := b64.StdEncoding.EncodeToString([]byte(username))
			passwordEnc := b64.StdEncoding.EncodeToString([]byte(password))
			loginMessage := "login " + usernameEnc + " " + passwordEnc + "\n"
			fmt.Fprintf(conn, loginMessage)

			res, err := serverReader.ReadString('\n')
			errorHandler(err)
			res = strings.Replace(res, "\n", "", -1)
			if res == "no" {
				fmt.Println("username not found / password incorrect")
			} else if res == "yes" {
				user = username
				break
			}
		} else if opt == "2" {
			// register: input username and password
			fmt.Print("register: \nusername: ")
			username, err := stdinReader.ReadString('\n')
			username = strings.Replace(username, "\n", "", -1)
			errorHandler(err)

			fmt.Print("password: ")
			password, err := stdinReader.ReadString('\n')
			password = strings.Replace(password, "\n", "", -1)
			errorHandler(err)

			usernameEnc := b64.StdEncoding.EncodeToString([]byte(username))
			passwordEnc := b64.StdEncoding.EncodeToString([]byte(password))
			loginMessage := "register " + usernameEnc + " " + passwordEnc + "\n"
			fmt.Fprintf(conn, loginMessage)
			res, err := serverReader.ReadString('\n')
			errorHandler(err)
			res = strings.Replace(res, "\n", "", -1)
			if res == "no" {
				fmt.Println("username is in use")
			} else if res == "yes" {
				user = username
				break
			}
		} else {
			// invalid option
			fmt.Println("option not found")
			continue
		}
	}

	// home page, select which option to do
	var friend string
	_ = friend

outerLoop:
	for {
		fmt.Println("Home\n (1) List all friends\n (2) Add friend\n (3) Delete a friend\n (4) Choose a chat room\n (5) Exit")
		opt, err := stdinReader.ReadString('\n')
		errorHandler(err)
		opt = strings.Replace(opt, "\n", "", -1)
		switch opt {
		case "1":
			fmt.Fprintf(conn, "listFriends\n")
			res, err := serverReader.ReadString('\n')
			errorHandler(err)
			res = strings.Replace(res, "\n", "", -1)
			friendList := strings.Split(res, " ")
			for i := 0; i < len(friendList); i++ {
				friendDec, err := b64.StdEncoding.DecodeString(friendList[i])
				errorHandler(err)
				fmt.Print(string(friendDec) + " ")
			}
			fmt.Println()
		case "2":
			fmt.Print("enter friend name you want to add: ")
			friendAdd, err := stdinReader.ReadString('\n')
			friendAdd = strings.Replace(friendAdd, "\n", "", -1)
			errorHandler(err)
			friendAddEnc := b64.StdEncoding.EncodeToString([]byte(friendAdd))

			cmd := "addFriend " + friendAddEnc + "\n"
			fmt.Fprintf(conn, cmd)

			res, err := serverReader.ReadString('\n')
			errorHandler(err)
			res = strings.Replace(res, "\n", "", -1)

			if res == "ok" {
				fmt.Println("friend add")
			} else if res == "added" {
				fmt.Println("friend already added")
			} else if res == "nonexist" {
				fmt.Println("user not exist")
			}
		case "3":
			fmt.Print("enter friend name you want to delete: ")
			friendDel, err := stdinReader.ReadString('\n')
			friendDel = strings.Replace(friendDel, "\n", "", -1)
			errorHandler(err)
			friendDelEnc := b64.StdEncoding.EncodeToString([]byte(friendDel))
			cmd := "deleteFriend " + friendDelEnc + "\n"
			fmt.Fprintf(conn, cmd)

			res, err := serverReader.ReadString('\n')
			errorHandler(err)
			res = strings.Replace(res, "\n", "", -1)

			if res == "ok" {
				fmt.Println("friend deleted")
			} else if res == "failed" {
				fmt.Println("invald username")
			}
		case "4":
			fmt.Fprintf(conn, "listChatroom\n")
			res, err := serverReader.ReadString('\n')
			errorHandler(err)
			res = strings.Replace(res, "\n", "", -1)

			chatroomNum, err := strconv.Atoi(res)
			if chatroomNum == 0 {
				fmt.Println("chatroom list is empty")
			} else {
				for i := 0; i < chatroomNum; i++ {
					res, err := serverReader.ReadString('\n')
					errorHandler(err)
					res = strings.Replace(res, "\n", "", -1)
					chatroomInfo := strings.Split(res, " ")
					mem1, err := b64.StdEncoding.DecodeString(chatroomInfo[1])
					errorHandler(err)
					mem2, err := b64.StdEncoding.DecodeString(chatroomInfo[2])
					errorHandler(err)
					fmt.Printf("(" + chatroomInfo[0] + ") ")
					if user == string(mem1) {
						fmt.Println(string(mem2))
					} else {
						fmt.Println(string(mem1))
					}
				}
			}

			fmt.Println("Join a chatroom by ID or type c to create a chatroom")
			opt, err := stdinReader.ReadString('\n')
			opt = strings.Replace(opt, "\n", "", -1)
			if opt == "c" {
				// create a chatroom
				fmt.Print("type the friend name you want to add: ")
				friend, err := stdinReader.ReadString('\n')
				errorHandler(err)
				friend = strings.Replace(friend, "\n", "", -1)
				friendEnc := b64.StdEncoding.EncodeToString([]byte(friend))
				cmd := "createChatroom " + friendEnc + "\n"
				fmt.Fprintf(conn, cmd)
				res, err := serverReader.ReadString('\n')
				res = strings.Replace(res, "\n", "", -1)
				data := strings.Split(res, " ")
				if data[0] == "ok" {
					fmt.Println("chatroom created, ID: " + data[1])
				} else if data[0] == "failed" {
					fmt.Println("invalid username")
				}
			} else {
				// join chatroom by ID
				cmd := "joinChatroom " + opt + "\n"
				fmt.Fprintf(conn, cmd)
				// receive id member1 member2
				res, err := serverReader.ReadString('\n')
				errorHandler(err)
				res = strings.Replace(res, "\n", "", -1)
				chatroomInfo := strings.Split(res, " ")

				if chatroomInfo[0] == "ok" {
					mem1, err := b64.StdEncoding.DecodeString(chatroomInfo[1])
					errorHandler(err)
					mem2, err := b64.StdEncoding.DecodeString(chatroomInfo[2])
					errorHandler(err)
					if user == string(mem1) {
						friend = string(mem2)
					} else {
						friend = string(mem1)
					}
					break outerLoop
				} else {
					fmt.Println("invalid ID")
				}
			}

		case "5":
			fmt.Println("Bye~")
			os.Exit(0)
		default:
			fmt.Println("option not found")
			continue
		}
		// just for not showing error message in vs code, should be commented when uploading
		// break
	}

	// chatroom init
	// format: <left> <right>\n<user>:<msg>\n<user>:<msg>\n...
	fileMap := make(map[string]string)
	imageMap := make(map[string]string)
	tail := printLogs(conn, serverReader, fileMap, imageMap, 0)
	// inside chatroom
	fmt.Print("Chatroom option: \n (0) help\n (1) send message\n (2) send image\n (3) send file\n (4) refresh message\n (5) get image\n (6) get file\n (7) exit chatroom\n ")
	// make a channel, for non-blocking message sending
	ch := make(chan string, 10)

	createDirectory("./clientDir")
chatRoomLoop:
	for {
	LOOP:
		for {
			select {
			case v := <-ch:
				fmt.Print(v)
			default:
				break LOOP
			}
		}
		fmt.Print("opt: ")
		opt, err := stdinReader.ReadString('\n')
		errorHandler(err)
		opt = strings.Replace(opt, "\n", "", -1)
		switch opt {
		case "0":
			fmt.Println("Chatroom option: \n (0) help\n (1) send message\n (2) send image\n (3) send file\n (4) refresh message\n (5) get image\n (6) get file\n (7) exit chatroom")
		case "1":
			fmt.Printf("%s: ", user)
			msg, err := stdinReader.ReadString('\n')
			errorHandler(err)
			msg = strings.Replace(msg, "\n", "", -1)
			msgEnc := b64.StdEncoding.EncodeToString([]byte(msg))
			fmt.Fprintf(conn, "sendMessage %s\n", msgEnc)

		case "2":
			upload("Image", conn, stdinReader, serverReader, ch, wg)
		case "3":
			upload("File", conn, stdinReader, serverReader, ch, wg)
		case "4":
			tail = printLogs(conn, serverReader, fileMap, imageMap, tail)
		case "5":
			download("Image", conn, stdinReader, serverReader, ch, wg, imageMap)
		case "6":
			download("File", conn, stdinReader, serverReader, ch, wg, fileMap)
		case "7":
			fmt.Print("exit chatroom")
			break chatRoomLoop
		default:
			fmt.Println("invalid option")
		}
	}
	fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
	status, err := bufio.NewReader(conn).ReadString('\n')
	fmt.Println(status)
	wg.Wait()
}

func errorHandler(err error) {
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

func createDirectory(dir string) {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err := os.Mkdir(dir, os.ModePerm)
		errorHandler(err)
	} else if err == nil {
		return
	} else {
		errorHandler(err)
	}
}

func upload(filetype string, conn net.Conn, stdinReader *bufio.Reader, serverReader *bufio.Reader, ch chan string, wg sync.WaitGroup) {
	fmt.Printf("%s name:", filetype)
	filename, err := stdinReader.ReadString('\n')
	errorHandler(err)
	filename = strings.Replace(filename, "\n", "", -1)
	file, err := os.Open("./clientDir/" + filename)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fmt.Fprintf(conn, "send%s\n", filetype)
	res, err := serverReader.ReadString('\n')
	errorHandler(err)
	res = strings.Replace(res, "\n", "", -1)
	data := strings.Split(res, " ")
	if data[0] == "ok" {
		token := data[1]
		wg.Add(1)
		go func() {
			defer wg.Done()
			fileConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", os.Args[1], os.Args[2]))
			errorHandler(err)
			filenameEnc := b64.StdEncoding.EncodeToString([]byte(filename))
			fi, err := file.Stat()
			errorHandler(err)

			fmt.Fprintf(fileConn, "upload%s %s %s %d\n", filetype, token, filenameEnc, fi.Size())
			// fmt.Printf("file %s %s %d\n", token, filenameEnc, fi.Size())
			fileServerReader := bufio.NewReader(fileConn)
			res, err := fileServerReader.ReadString('\n')
			errorHandler(err)
			res = strings.Replace(res, "\n", "", -1)

			if res == "ok" {
				_, err = io.Copy(fileConn, file)
				errorHandler(err)
				res, err := fileServerReader.ReadString('\n')
				errorHandler(err)
				res = strings.Replace(res, "\n", "", -1)
				if res == "ok" {
					ch <- fmt.Sprintf("%s %s sent\n", filetype, filename)
				} else {
					ch <- fmt.Sprintf("%s %s failed to send\n", filetype, filename)
				}
				fileConn.Close()
			} else {
				ch <- "authentication failed\n"
				fileConn.Close()
				return
			}
		}()
	} else {
		fmt.Println("coommand failed")
	}
}

func download(filetype string, conn net.Conn, stdinReader *bufio.Reader, serverReader *bufio.Reader, ch chan string, wg sync.WaitGroup, fileMap map[string]string) {
	numMap := make(map[int]string)
	fmt.Fprintf(conn, "list%ss\n", filetype)
	res, err := serverReader.ReadString('\n')
	errorHandler(err)
	res = strings.Replace(res, "\n", "", -1)
	num, err := strconv.Atoi(res)
	if num == 0 {
		fmt.Println("No files available")
		return
	}
	for i := 0; i < num; i++ {
		res, err := serverReader.ReadString('\n')
		errorHandler(err)
		res = strings.Replace(res, "\n", "", -1)
		data := strings.Split(res, " ")
		filename, err := b64.StdEncoding.DecodeString(data[0])
		token := data[1]
		fileMap[token] = string(filename)
		numMap[i] = token
		fmt.Printf("(%d) %s ", i, string(filename))
	}
	fmt.Printf("\nChoose the %s number you want to download: ", filetype)
	res, err = stdinReader.ReadString('\n')
	errorHandler(err)
	res = strings.Replace(res, "\n", "", -1)
	which, err := strconv.Atoi(res)
	if which >= num {
		fmt.Println("Index out of range")
		return
	}
	errorHandler(err)
	// open a new thread
	wg.Add(1)
	go func() {
		defer wg.Done()
		fileConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", os.Args[1], os.Args[2]))
		fileServerReader := bufio.NewReader(fileConn)
		errorHandler(err)
		fmt.Fprintf(fileConn, "download%s %s\n", filetype, numMap[which])
		res, err = fileServerReader.ReadString('\n')
		errorHandler(err)
		res = strings.Replace(res, "\n", "", -1)
		data := strings.Split(res, " ")
		if data[0] == "ok" {
			filename, err := b64.StdEncoding.DecodeString(data[1])
			errorHandler(err)
			filesize, err := strconv.Atoi(data[2])
			errorHandler(err)
			filepath := "./clientDir/" + string(filename)
			file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
			errorHandler(err)
			sz, err := io.Copy(file, fileServerReader)
			errorHandler(err)
			if sz != int64(filesize) {
				ch <- fmt.Sprintf("%s copy incomplete\n", filetype)
				return
			} else {
				ch <- fmt.Sprintf("%s %s received\n", filetype, string(filename))
			}
			file.Close()
		}
		fileConn.Close()
	}()
}

func printLogs(conn net.Conn, serverReader *bufio.Reader, fileMap map[string]string, imageMap map[string]string, start int) int {
	fmt.Fprintf(conn, "logs %d -1\n", start)
	res, err := serverReader.ReadString('\n')
	errorHandler(err)
	res = strings.Replace(res, "\n", "", -1)
	border := strings.Split(res, " ")
	left, err := strconv.Atoi(border[0])
	errorHandler(err)
	right, err := strconv.Atoi(border[1])
	errorHandler(err)
	for i := left; i < right; i++ {
		res, err := serverReader.ReadString('\n')
		errorHandler(err)
		res = strings.Replace(res, "\n", "", -1)
		data := strings.Split(res, " ")
		from, err := b64.StdEncoding.DecodeString(data[0])
		errorHandler(err)
		dataType := data[1]
		msg, err := b64.StdEncoding.DecodeString(data[2])
		errorHandler(err)
		if dataType == "text" {
			fmt.Printf("%s:%s\n", string(from), string(msg))
		} else if dataType == "file" {
			fileMap[data[3]] = string(msg)
			fmt.Printf("%s:[%s]\n", string(from), string(msg))
		} else if dataType == "image" {
			imageMap[data[3]] = string(msg)
			fmt.Printf("%s:[%s]\n", string(from), string(msg))
		}
	}
	return right
}
