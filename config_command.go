package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"strings"
	"time"

	"github.com/Felyne/config_center"
	"github.com/coreos/etcd/clientv3"
)

var dialTimeout = 15 * time.Second

type ConfigCommand struct {
}

func (cm *ConfigCommand) CommandName() string {
	return "config"
}

func (cm *ConfigCommand) Run() error {
	if len(os.Args) < 3 {
		fmt.Printf("Usage:%s %s [subCommand] [args]\n", os.Args[0], cm.CommandName())
		fmt.Printf("    subCommand are:get set dump restore\n")
		os.Exit(1)
	}
	subCommand := os.Args[2]
	args := os.Args[3:]
	if subCommand == "set" {
		return cm.runSet(args)
	} else if subCommand == "get" {
		return cm.runGet(args)
	} else if subCommand == "dump" {
		return cm.runDump(args)
	} else if subCommand == "restore" {
		return cm.runRestore(args)
	}
	return nil
}

func (cm *ConfigCommand) runSet(args []string) error {
	if len(args) < 4 {
		fmt.Printf("Usage:%s %s set [envName] [cfgName] [fileName] [etcdAddr...]\n",
			os.Args[0], cm.CommandName())
		os.Exit(1)
	}
	envName := args[0]
	cfgName := args[1]
	fileName := args[2]
	etcdAddrs := args[3:]

	f, err := os.Open(fileName)
	if nil != err {
		return err
	}
	defer f.Close()

	content, err := ioutil.ReadAll(f)
	if nil != err {
		return err
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdAddrs,
		DialTimeout: dialTimeout,
	})
	if nil != err {
		return err
	}
	cc := config_center.NewConfigCenter(cli, envName)
	err = cc.SetConfig(cfgName, string(content))
	return nil
}

func (cm *ConfigCommand) runGet(args []string) error {
	if len(args) < 3 {
		fmt.Printf("Usage:%s %s get [envName] [cfgName] [etcdAddr...]\n",
			os.Args[0], cm.CommandName())
		os.Exit(1)
	}
	envName := args[0]
	cfgName := args[1]
	etcdAddrs := args[2:]
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdAddrs,
		DialTimeout: dialTimeout,
	})
	if nil != err {
		return err
	}
	cc := config_center.NewConfigCenter(cli, envName)
	content, err := cc.GetConfig(cfgName)
	if nil != err {
		return err
	}
	fmt.Println(content)
	return nil
}

func (cm *ConfigCommand) runDump(args []string) error {
	if len(args) < 3 {
		fmt.Printf("Usage:%s %s dump [envName] [dirPath] [etcdAddr...]\n",
			os.Args[0], cm.CommandName())
		os.Exit(1)
	}
	envName := args[0]
	dirPath := args[1]
	etcdAddrs := args[2:]

	_ = os.MkdirAll(dirPath, 0700)

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdAddrs,
		DialTimeout: dialTimeout,
	})
	if nil != err {
		return err
	}
	cc := config_center.NewConfigCenter(cli, envName)
	cfgMap, err := cc.ListConfig()
	if nil != err {
		return nil
	}
	for cfgName, content := range cfgMap {
		func() {
			fileName := strings.Join([]string{dirPath, cfgName},
				config_center.PathSeparator)
			_ = ioutil.WriteFile(fileName, []byte(content), 0600)
			fmt.Printf("dump %s\n", cfgName)
		}()
	}
	return nil
}

func (cm *ConfigCommand) runRestore(args []string) error {
	if len(args) < 3 {
		fmt.Printf("Usage:%s %s restore [envName] [dirPath] [etcdAddr...]\n",
			os.Args[0], cm.CommandName())
		os.Exit(1)
	}
	envName := args[0]
	dirPath := args[1]
	etcdAddrs := args[2:]

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdAddrs,
		DialTimeout: dialTimeout,
	})
	if nil != err {
		return err
	}
	cc := config_center.NewConfigCenter(cli, envName)

	dir, err := os.Open(dirPath)
	if nil != err {
		return err
	}
	files, err := dir.Readdir(0)
	if nil != err {
		return err
	}
	for _, file := range files {
		f, err := os.Open(strings.Join([]string{dirPath, file.Name()},
			config_center.PathSeparator))
		if nil != err {
			fmt.Println(err)
			continue
		}
		content, err := ioutil.ReadAll(f)
		if nil != err {
			fmt.Println(err)
			continue
		}
		_ = cc.SetConfig(file.Name(), string(content))
		fmt.Printf("restore %s\n", file.Name())
	}
	return nil
}