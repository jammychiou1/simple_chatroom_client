package main

import (
    "net"
    "os"
    "fmt"
    "bufio"
    "strings"
    "encoding/base64"
    "strconv"
    "io"
    "log"
)

type webClient struct {
    r *bufio.Reader
    w *bufio.Writer
    conn net.Conn

    path []string
    method string
    body []byte
    username string
    password string
    contentLength int

    rspHeaders map[string]string
}

func statusCodeText(statusCode int) string {
    if statusCode == 200 {
        return "OK"
    }
    if statusCode == 400 {
        return "Bad Request"
    }
    return ""
}

func parseInfo(clnt *webClient) (ok bool) {
    ln, err := clnt.r.ReadString('\n')
    if err != nil {
        return false
    }
    tokens := strings.Split(ln, " ")
    if len(tokens) < 3 {
        return false
    }
    clnt.method = tokens[0]
    clnt.path = strings.Split(tokens[1], "/")
    clnt.contentLength = 0
    for {
        ln, err := clnt.r.ReadString('\n')
        if err != nil {
            return false
        }
        ln = strings.TrimSpace(ln)
        log.Println(ln)
        key_val := strings.Split(ln, ": ")
        if key_val[0] == "Authorization" {
            log.Println("is auth")
            log.Println(len(key_val))
            if len(key_val) != 2 {
                return false
            }
            log.Printf("\"%s\"\n", key_val[1])
            basic_auth := strings.Split(string(key_val[1]), " ")
            log.Println(len(basic_auth))
            if len(basic_auth) != 2 {
                return false
            }
            auth := basic_auth[1]
            dec, err := base64.StdEncoding.DecodeString(auth)
            log.Println(dec)
            if err != nil {
                return false
            }
            // TODO disallow colon in username and password
            user_pass := strings.Split(string(dec), ":")
            if len(user_pass) != 2 {
                return false
            }
            clnt.username = user_pass[0]
            clnt.password = user_pass[1]
        }
        if key_val[0] == "Content-Length" {
            if len(key_val) != 2 {
                return false
            }
            clnt.contentLength, err = strconv.Atoi(key_val[1])
            if err != nil {
                return false
            }
        }
        if ln == "" {
            break
        }
    }
    if clnt.contentLength > 0 {
        bodyBytes := make([]byte, clnt.contentLength)
        n, err := io.ReadFull(clnt.r, bodyBytes)
        if err != nil {
            log.Println(err)
            return false
        }
        if n != clnt.contentLength {
            log.Println("body too short")
            return false
        }
        clnt.body = bodyBytes
    }
    return true
}

func writeHeader(clnt webClient, statusCode int, rspContentLength int) {
    fmt.Fprintf(clnt.w, "HTTP/1.0 %d %s\n", statusCode, statusCodeText(statusCode))
    for k, v := range clnt.rspHeaders {
        fmt.Fprintf(clnt.w, "%s: %s\n", k, v)
    }
    fmt.Fprintf(clnt.w, "Content-Length: %d\n\n", rspContentLength)
    clnt.w.Flush()
}

func handleRequest(clnt webClient) {
    defer clnt.conn.Close()
    if !parseInfo(&clnt) {
        writeHeader(clnt, 400, 0)
        return
    }
    log.Printf("body %q\n", clnt.body)
    log.Printf("method %s\n", clnt.method)
    log.Printf("path %v\n", clnt.path)
    
    if len(clnt.path) == 2 {
        if clnt.path[1] == "register" {
            handleRegister(clnt)
        } else if clnt.path[1] == "check" {
            handleCheck(clnt)
        } else if clnt.path[1] == "chatrooms" {
            if clnt.method == "GET" {
                handleListChatrooms(clnt)
            } else {
                handleAddChatrooms(clnt)
            }
        } else if clnt.path[1] == "friends" {
            if clnt.method == "GET" {
                handleListFriends(clnt)
            } else if clnt.method == "DELETE" {
                handleDeleteFriend(clnt)
            } else {
                handleAddFriend(clnt)
            }
        } else {
            writeHeader(clnt, 400, 0)
            return
        }
    } else if len(clnt.path) == 3 {
        if clnt.path[1] == "static" {
            handleStatic(clnt)
        } else if clnt.path[1] == "files" {
            handleDownloadFile(clnt)
        } else if clnt.path[1] == "images" {
            handleDownloadImage(clnt)
        } else {
            writeHeader(clnt, 400, 0)
            return
        }
    } else if len(clnt.path) == 4 {
        if clnt.path[1] == "chatrooms" {
            handleChatrooms(clnt)
        } else {
            writeHeader(clnt, 400, 0)
            return
        }
    } else {
        writeHeader(clnt, 400, 0)
        return
    }
}

func runWebserver() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Args[3]))
    errorHandler(err)
    defer listener.Close()
    for {
        conn, err := listener.Accept()
        log.Println("new web client")
        errorHandler(err)
        clnt := webClient{conn: conn}    
        clnt.r = bufio.NewReader(clnt.conn)
        clnt.w = bufio.NewWriter(clnt.conn)
        clnt.rspHeaders = make(map[string]string)
        go handleRequest(clnt)
    }
}
