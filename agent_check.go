package main

import (
	"bufio"
	"io"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	systemstat "bitbucket.org/bertimus9/systemstat"
)

//TODO this should NOT be a global
var CommandStr string
var IdleInt int = 88

func main() {
	command := make(chan string, 1)
	CommandStr = "UP"
	ln, err := net.Listen("tcp", ":5309")
	if err != nil {
		log.Fatalln("there was an error:", err)
	}
	go Talk(ln, command)

	ln2, err := net.Listen("tcp", "localhost:8675")
	if err != nil {
		log.Fatalln("there was an error:", err)
	}
	go Listen(ln2, command)
	go updateIdle()
	go updateCommand(command)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	s := <-c
	log.Println("exiting on:", s)
}

func updateCommand(command chan string) {
	for {
		CommandStr = <-command
	}
}

func updateIdle() {
	var sample [2]systemstat.CPUSample
	for {
		for i := 0; i < 2; i++ {
			sample[i] = systemstat.GetCPUSample()
			time.Sleep(5000 * time.Millisecond)
		}
		avg := systemstat.GetSimpleCPUAverage(sample[0], sample[1])
		idlePercent := avg.IdlePct
		IdleInt = int(idlePercent)
		fmt.Println("idleInt:", IdleInt, "sample0:", int(sample[0].User), "sample1:", int(sample[1].User))
	}
}

func get_idle() (out int) {
	return int(IdleInt)
}

func handleTalk(conn net.Conn, command <-chan string) {
	defer conn.Close()
	idle := strconv.Itoa(get_idle())
	io.WriteString(conn, CommandStr+" "+idle+"% \n")
	return
}

func handleListen(conn net.Conn, command chan string) {
	defer conn.Close()
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return
	}
	line = strings.Replace(line, "\n", "", -1)
	command <- line
	conn.Write([]byte(line + " OK \n"))
	return
}

func Talk(ln net.Listener, command chan string) {
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("there was an error:", err)
			break
		}
		go handleTalk(conn, command)
	}
}

func Listen(ln net.Listener, command chan string) {
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("there was an error:", err)
			break
		}
		go handleListen(conn, command)
	}

}
