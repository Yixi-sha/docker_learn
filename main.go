package main

import (
	"log"
	"github.com/urfave/cli"
	"os"
	"fmt"
)


func main(){
	app := cli.NewApp()
	app.Name = "mydocker"
	app.Usage = usage

	fmt.Println(os.Args)
	app.Commands = []*cli.Command{
		&initCommand,
		&runCommand,
	}
	
	app.Before = func(context *cli.Context) error{
		log.Println("before")
		return nil
	}

	if err := app.Run(os.Args); err != nil{
		log.Fatal(err)
	}
}