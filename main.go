package main

import (
	"bufio"
	b64 "encoding/base64"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Println(os.Args[1])
	fmt.Println(fmt.Sprintf("%s:%s", os.Args[1], os.Args[2]))
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", os.Args[1], os.Args[2]))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

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
			for {
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
					break
				}
			}
		} else if opt == "2" {
			// register: input username and password
			for {
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
					break
				}
			}
		} else {
			// invalid option
			fmt.Println("option not found")
			continue
		}
		break
	}

	// home page, select which option to do
	for {
		fmt.Println("Home\n (1) List all friends\n (2) Add friend\n (3) Delete a friend\n (4) Choose a chat room\n (5) Exit")
		opt, err := stdinReader.ReadString('\n')
		errorHandler(err)
		opt = strings.Replace(opt, "\n", "", -1)
		switch opt {
		case "1":
			fmt.Fprintf(conn, "list\n")
			res, err := serverReader.ReadString('\n')
			errorHandler(err)
			fmt.Print(res)
		case "2":
			fmt.Print("enter friend name you want to add: ")
			friendAdd, err := stdinReader.ReadString('\n')
			errorHandler(err)
			cmd := "add " + friendAdd
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
			errorHandler(err)
			cmd := "delete " + friendDel
			fmt.Fprintf(conn, cmd)
			fmt.Fprintf(conn, "delete\n")

			res, err := serverReader.ReadString('\n')
			errorHandler(err)
			res = strings.Replace(res, "\n", "", -1)
			if res == "ok" {
				fmt.Println("friend deleted")
			} else if res == "failed" {
				fmt.Println("invald username")
			}
		case "4":
			fmt.Fprintf(conn, "choose\n")
			chatList, err := serverReader.ReadString('\n')
			errorHandler(err)
			fmt.Print(chatList)

			target, err := stdinReader.ReadString('\n')
			errorHandler(err)
			cmd := "choose " + target
			fmt.Fprintf(conn, cmd)
			res, err := serverReader.ReadString('\n')
			errorHandler(err)
			res = strings.Replace(res, "\n", "", -1)
			if res == "ok" {
				break
			} else {
				fmt.Println("invalid username")
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
	for {

	}
	fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
	status, err := bufio.NewReader(conn).ReadString('\n')
	fmt.Println(status)
}

func errorHandler(err error) {
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

func home() {

}
