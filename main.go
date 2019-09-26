package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type Command interface {
	CommandName() string
	Run() error
}

var (
	Version   = ""
	BuildTime = ""
)
var getVersion = flag.Bool("v", false, "print current version")

var commandList []Command

func main() {
	flag.Parse()
	if *getVersion {
		fmt.Println("Version:", Version)
		fmt.Println("BuildTime:", BuildTime)
		return
	}
	commandList := []Command{
		&ConfigCommand{},
	}
	if len(os.Args) < 2 {
		help()
	}
	for _, cmd := range commandList {
		if cmd.CommandName() == os.Args[1] {
			err := cmd.Run()
			if nil != err {
				fmt.Println(err)
			}
		}
	}
}

func help() {
	var nameList []string
	for _, cmd := range commandList {
		nameList = append(nameList, cmd.CommandName())
	}
	fmt.Printf("Usage:%s [command] [args]\n", os.Args[0])
	fmt.Printf("    commands are:%s\n", strings.Join(nameList, ","))
	os.Exit(1)
}
