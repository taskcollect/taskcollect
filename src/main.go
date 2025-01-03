package main

import (
	"flag"

	"main/logger"
	"main/server"
)

var wflag bool

var schools = []string{
	"example",
	"gihs",
	"uofa",
}

func init() {
	flag.BoolVar(&wflag, "w", false, "run without TLS, on port 8080")
}

func main() {
	flag.Parse()
	server.Enrol(schools)
	err := server.Configure(version)
	if err != nil {
		logger.Fatal(err)
	}
	err = server.Run(!wflag)
	if err != nil {
		logger.Fatal(err)
	}
}
