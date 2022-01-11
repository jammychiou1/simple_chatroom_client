package main

import (
    "fmt"
    "net"
    "bufio"
    "os"
)

func main() {
    fmt.Println(os.Args[1])
    fmt.Println(fmt.Sprintf("%s:%s", os.Args[1], os.Args[2]))
    conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", os.Args[1], os.Args[2]))
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
    status, err := bufio.NewReader(conn).ReadString('\n')
    fmt.Println(status)
}
