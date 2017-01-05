// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

package main

import (
	"errors"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path"
	"syscall"
	"time"

	"github.com/Supernomad/quantum/agg"
	"github.com/Supernomad/quantum/common"
	"github.com/Supernomad/quantum/datastore"
)

func handleError(log *common.Logger, err error) {
	if err != nil {
		log.Error.Println(err.Error())
		os.Exit(1)
	}
}

func main() {
	log := common.NewLogger(common.DebugLogger)

	cfg, err := common.NewConfig(log)
	handleError(log, err)

	cfgSrv := &common.ConfigServer{
		Config: cfg,
	}

	store, err := datastore.New(datastore.ETCDDatastore, log, cfg)
	handleError(log, err)

	err = store.Init()
	handleError(log, err)

	storeSrv := &datastore.DatastoreServer{
		Store: store,
	}

	aggregator := agg.New(log, cfg)

	aggSrv := &agg.AggServer{
		Agg: aggregator,
	}

	aggregator.Start()
	store.Start()

	rpc.Register(storeSrv)
	rpc.Register(aggSrv)
	rpc.Register(cfgSrv)

	rpc.HandleHTTP()

	sockPath := path.Join(cfg.DataDir, "quantum.sock")

	os.Remove(sockPath)
	l, err := net.ListenUnix("unix", &net.UnixAddr{Name: sockPath})
	if err != nil {
		handleError(log, errors.New("error opening rpc tunnel: "+err.Error()))
	}
	go http.Serve(l, nil)

	files := make([]uintptr, 3)
	files[0] = os.Stdin.Fd()
	files[1] = os.Stdout.Fd()
	files[2] = os.Stderr.Fd()

	args := make([]string, 2)
	args[0] = path.Join(path.Dir(os.Args[0]), "quantum-worker")
	args[1] = sockPath

	for i := 0; i < cfg.NumWorkers; i++ {
		syscall.ForkExec(args[0], args, &syscall.ProcAttr{Env: os.Environ(), Files: files})
		time.Sleep(10 * time.Second)
	}

	log.Info.Println("[MASTER]", "[MAIN]", "Start up complete.")

	signaler := common.NewSignaler(log, cfg, []int{}, map[string]string{})

	err = signaler.Wait(true)
	handleError(log, err)

	aggregator.Stop()
	store.Stop()
}
