package server

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func logStream(w http.ResponseWriter, r *http.Request) {

	flusher, ok := w.(http.Flusher)
	if !ok {
		panic("expected http.ResponseWriter to be an http.Flusher")
	}

	w.Header().Set("Connection", "Keep-Alive")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)

	scaner := bufio.NewScanner(r.Body)
	for scaner.Scan() {
		text := scaner.Text()
		fmt.Fprintf(w, "time: %s, data: %s\n", time.UnixDate, text)
		flusher.Flush()
		println(text)
	}

}

func Listen(addr string) error {

	mux := http.NewServeMux()
	mux.HandleFunc("/log/stream", logStream)

	srv := http.Server{
		Addr:           addr,
		Handler:        mux,
		ReadTimeout:    time.Second * 10,
		WriteTimeout:   time.Second * 10,
		MaxHeaderBytes: 10240,
	}

	return listenServer(srv)

}

func listenServer(srv http.Server) error {

	var proto string

	if srv.Addr == "" {
		srv.Addr = ":http"
	}

	if strings.Contains(srv.Addr, "/") {
		proto = "unix"
		err := checkUnixSocketFile(srv.Addr)
		if err != nil {
			return err
		}
		autoRemoveUnixSocketFile(srv.Addr)
	} else {
		proto = "tcp"
	}

	l, e := net.Listen(proto, srv.Addr)
	if e != nil {
		return e
	}

	return srv.Serve(l)

}

func checkUnixSocketFile(file string) error {

	exists, err := checkFileExists(file)
	if err != nil {
		return err
	}
	if exists {
		return errors.New(fmt.Sprintf("unix file %s is already exists", file))
	}

	return nil

}

func checkFileExists(file string) (bool, error) {

	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, err
		}
	} else {
		return true, nil
	}

}

func autoRemoveUnixSocketFile(file string) {

	sig := make(chan os.Signal, 1)
	signal.Notify(sig,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT)

	removeFile := func(s os.Signal) {
		log.Printf("get signal %s, trying to remove unix socket file %s before process exit\n", s, file)
		err := os.Remove(file)
		if err != nil {
			log.Printf("fail to remove file: %s\n", err)
		}
		log.Printf("exit\n")
		os.Exit(0)
	}

	go func() {
		s := <-sig
		switch s {
		case os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT:
			removeFile(s)
		}
	}()

}