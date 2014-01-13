package main

import (
    "strings"
    "net"
    "fmt"
    "os"
    "bufio"
    "io"
    "log"
)

func main() {
    conn, err := net.Dial("tcp", ":6000")

    serverchan := make(chan string)
    stdinchan := make(chan string)

    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("New connection established\n")

    go WriteToServer(conn, serverchan)
    go WriteToStdin(stdinchan)
    go ReadFromServer(conn, stdinchan, serverchan)

    reader := bufio.NewReader(os.Stdin)

    for {
        fmt.Print(">>> ")
        line, err := reader.ReadString('\n')
        if err != nil {
            fmt.Println("Read error")
        }

        serverchan <- line
    }
}

func WriteToStdin(stdinchan <-chan string) {
    for msg := range stdinchan {
        fmt.Println(msg)
    }
}

func WriteToServer(c net.Conn, serverchan <-chan string) {
    for msg := range serverchan {
        io.WriteString(c, msg)
    }
}

func promptNick() string {
    reader := bufio.NewReader(os.Stdin)

    fmt.Print("nick>> ")
    line, err := reader.ReadString('\n')
    if err != nil {
        fmt.Println("Read error")
    }
    return line
}

func ReadFromServer(c net.Conn, stdinchan chan<- string, serverchan chan<- string) {
    bufc := bufio.NewReader(c)
    for {
        line, err := bufc.ReadString('\n')
        if err != nil {
            break
        }
        if strings.EqualFold("What is your nick?\n", line) {
            stdinchan <- line
            serverchan <- promptNick()
        } else {
            stdinchan <- line
        }
    }
}
