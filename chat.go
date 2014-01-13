package main

import (
    "log"
    "net"
    "bufio"
    "io"
    "fmt"
    "strings"
)

type Client struct {
    nickname string
    conn net.Conn
    ch chan string
}

func main() {
    ln, err := net.Listen("tcp", ":6000")
    if err != nil {
        log.Fatal(err)
    }

    msgchan := make(chan string)
    addchan := make(chan Client)
    rmchan := make(chan Client)

    go handleMessage(msgchan, addchan, rmchan)

    for {
        conn, err := ln.Accept()
        if err != nil {
            log.Println(err)
            continue
        }
        go handleConnection(conn, msgchan, addchan, rmchan)
    }
}

func handleConnection(c net.Conn, msgchan chan<- string, addchan chan<- Client, rmchan chan<- Client) {
    bufc := bufio.NewReader(c)
    defer c.Close()
    client := Client{
        conn     : c,
        nickname : promptNick(c, bufc),
        ch       : make(chan string),
    }

    if strings.TrimSpace(client.nickname) == "" {
        io.WriteString(c, "Invalid Username\n")
        return
    }
    addchan <- client
    defer func() {
        msgchan <- fmt.Sprintf("User %s left the chat room. \n", client.nickname)
        log.Printf("Connection from %v closed.\n", c.RemoteAddr())
        rmchan <- client
    }()
    io.WriteString(c, fmt.Sprintf("Welcome, %s!\n\n", client.nickname))
    msgchan <- fmt.Sprintf("New user %s has joined the chat room. \n", client.nickname)

    go client.ReadLinesInto(msgchan)
    client.WriteLinesFrom(client.ch)
}

func promptNick(c net.Conn, bufc *bufio.Reader) string {
    io.WriteString(c, "\033[1;30;41mWelcome to the fancy demo chat!\033[0m\n")
    io.WriteString(c, "What is your nick?\n")
    nick, _, _ := bufc.ReadLine()
    return string(nick)
}

func (c Client) ReadLinesInto(ch chan<- string) {
    bufc := bufio.NewReader(c.conn)
    for {
        line, err := bufc.ReadString('\n')
        if err != nil {
            break
        }
        ch <- fmt.Sprintf("%s: %s", c.nickname, line)
    }
}

func (c Client) WriteLinesFrom(ch <-chan string) {
    for msg := range ch {
        _, err := io.WriteString(c.conn, msg)
        if err != nil {
            return
        }
    }
}

func handleMessage(msgchan <-chan string, addchan <-chan Client, rmchan <-chan Client) {
    clients := make(map[net.Conn]chan<- string)

    for {
        select {
            case msg := <-msgchan:
                for _, ch := range clients {
                    go func(mch chan<- string) { mch <- "\033[1;33;40m" + msg + "\033[m\r\n" }(ch)
                    //ch <- "\033[1;33;40m" + msg + "\033[m\r\n"
                }
            case client := <-addchan:
                clients[client.conn] = client.ch
            case conn := <-rmchan:
                delete(clients, conn.conn)
        }
    }
}
