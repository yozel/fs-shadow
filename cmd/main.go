package main

import (
	"github.com/ayhanozemre/fs-shadow/watcher"
	log "github.com/sirupsen/logrus"
)

func main() {
	//tw, err := watcher.NewFSPathWatcher("/home/wade/Desktop/TransferChain")
	tw, err := watcher.NewVirtualPathWatcher("/home/wade/Desktop/TransferChain")

	if err == nil {
		tw.PrintTree("INIT TREE")
		done := make(chan bool)
		<-done
	} else {
		log.Panic(err)
	}
	tw.Close()
}
