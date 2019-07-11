package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"os"
	"runtime"
	"time"
)

const VERSION = "v1"

func main() {
	app := cli.NewApp()
	app.Name = "pulse client"
	app.Usage = "a tool for driving pulse cli"
	app.Author = "Zhangjianxin"
	app.Email = "zhangjianxinnet@gmail.com"
	app.Version = fmt.Sprintf("%s %s/%s %s", VERSION,
		runtime.GOOS, runtime.GOARCH, runtime.Version())
	app.EnableBashCompletion = true
	app.Compiled = time.Now()

	app.Commands = []cli.Command{
		NewConfListCommand(),
		NewServiceListCommand(),
		NewAddTaskCommand(),
		NewCleanCommand(),
		NewHostsCommand(),
	}

	app.Run(os.Args)
}