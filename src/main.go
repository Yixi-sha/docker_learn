package main

import (
	"fmt"
	"log"
	"os"

	_ "mydocker/nsenter"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "mydocker"
	app.Usage = usage

	fmt.Println(os.Args)
	app.Commands = []cli.Command{
		initCommand,
		runCommand,
		commmitCommand,
		listCommand,
		logCommand,
		execCommand,
		stopCommand,
		removeCommand,
		networkCommand,
	}

	app.Before = func(context *cli.Context) error {
		log.Println("before")
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
