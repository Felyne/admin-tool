package main

type Command interface {
	CommandName() string
	Run() error
}
