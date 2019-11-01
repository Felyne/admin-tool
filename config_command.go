package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/Felyne/configcenter"
	"github.com/coreos/etcd/clientv3"
)

type ConfigCommand struct {
}

func (cm *ConfigCommand) CommandName() string {
	return "config"
}

func (cm *ConfigCommand) Run() error {
	if len(os.Args) < 3 {
		fmt.Printf("Usage:%s %s [subCommand] [args]\n", os.Args[0], cm.CommandName())
		fmt.Printf("    subCommand are: get set del dump restore\n")
		os.Exit(1)
	}
	subCommand := os.Args[2]
	args := os.Args[3:]
	if subCommand == "set" {
		return cm.runSet(args)
	} else if subCommand == "get" {
		return cm.runGet(args)
	} else if subCommand == "del" {
		return cm.runDel(args)
	} else if subCommand == "dump" {
		return cm.runDump(args)
	} else if subCommand == "restore" {
		return cm.runRestore(args)
	}
	return nil
}

//上传单个配置文件
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

	cli, err := getEtcdClient(etcdAddrs)
	if nil != err {
		return err
	}
	defer cli.Close()

	cc := configcenter.New(cli, envName)
	err = cc.SetConfig(cfgName, string(content))
	return nil
}

//获取单个配置信息
func (cm *ConfigCommand) runGet(args []string) error {
	if len(args) < 3 {
		fmt.Printf("Usage:%s %s get [envName] [cfgName] [etcdAddr...]\n",
			os.Args[0], cm.CommandName())
		os.Exit(1)
	}
	envName := args[0]
	cfgName := args[1]
	etcdAddrs := args[2:]

	cli, err := getEtcdClient(etcdAddrs)
	if nil != err {
		return err
	}
	defer cli.Close()

	cc := configcenter.New(cli, envName)
	content, err := cc.GetConfig(cfgName)
	if nil != err {
		return err
	}
	fmt.Println(content)
	return nil
}

//删除单个配置信息
func (cm *ConfigCommand) runDel(args []string) error {
	if len(args) < 3 {
		fmt.Printf("Usage:%s %s del [envName] [cfgName] [etcdAddr...]\n",
			os.Args[0], cm.CommandName())
		os.Exit(1)
	}
	envName := args[0]
	cfgName := args[1]
	etcdAddrs := args[2:]

	cli, err := getEtcdClient(etcdAddrs)
	if nil != err {
		return err
	}
	defer cli.Close()

	cc := configcenter.New(cli, envName)
	err = cc.RemoveConfig(cfgName)
	if nil != err {
		return err
	}
	fmt.Printf("del %s success\n", cfgName)
	return nil
}

//下载指定命名空间的配置到目标目录
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

	cli, err := getEtcdClient(etcdAddrs)
	if nil != err {
		return err
	}
	defer cli.Close()

	cc := configcenter.New(cli, envName)
	cfgMap, err := cc.ListConfig()
	if nil != err {
		return nil
	}
	for cfgName, content := range cfgMap {
		func() {
			fileName := strings.Join([]string{dirPath, cfgName},
				string(os.PathSeparator))
			_ = ioutil.WriteFile(fileName, []byte(content), 0600)
			fmt.Printf("dump %s\n", cfgName)
		}()
	}
	return nil
}

//上目录下的配置文件
func (cm *ConfigCommand) runRestore(args []string) error {
	if len(args) < 3 {
		fmt.Printf("Usage:%s %s restore [envName] [dirPath] [etcdAddr...]\n",
			os.Args[0], cm.CommandName())
		os.Exit(1)
	}
	envName := args[0]
	dirPath := args[1]
	etcdAddrs := args[2:]

	cli, err := getEtcdClient(etcdAddrs)
	if nil != err {
		return err
	}
	defer cli.Close()

	cc := configcenter.New(cli, envName)
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
			string(os.PathSeparator)))
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

func getEtcdClient(etcdAddrs []string) (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		Endpoints:   etcdAddrs,
		DialTimeout: 15 * time.Second,
	})
}
