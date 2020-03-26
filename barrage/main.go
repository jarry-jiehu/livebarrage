package main

import (
	log "code.google.com/p/log4go"
	"flag"
	"fmt"
	"livebarrage/comet"
	"livebarrage/conf"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"runtime/debug"
)

var (
	// Version information. Passed from "go build -ldflags"
	version     string
	versionInfo string
)

func showVersion() {
	fmt.Printf("%s\n", versionInfo)
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
		}
	}()

	var (
		//flVersion     = flag.Bool("v", false, "Print version information and quit")
		configFile    = flag.String("c", "./conf/conf.json", "Config file")
		logConfigFile = flag.String("l", "./conf/log4go.xml", "Log config file")
	)
	flag.Parse()

	//if *flVersion {
	//	showVersion()
	//	return
	//}

	cpus := runtime.NumCPU()
	runtime.GOMAXPROCS(cpus)

	err := conf.LoadConfig(*configFile)
	if err != nil {
		fmt.Printf("pushd: load config (%s) failed: (%s)\n", *configFile, err)
		os.Exit(1)
	}

	log.LoadConfiguration(*logConfigFile)

	cometServer := comet.InitServer(&conf.Config.Comet)
	log.Info("livebarrage: comet server inited")

	wg := &sync.WaitGroup{}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-c
		log.Info("pushd: received signal '%v', exiting\n", sig)
		///utils.RemovePidFile(srv.	runtime.config.Pidfile)
		cometServer.StopServer()
		//wg.Done()
		log.Info("pushd: leave 2")
	}()
	wg.Add(1)

	go func() {
		cometServer.StartServer()
	}()

	wg.Wait()
	log.Info("livebarrage: leave this cruel world!")
}
