package server

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
)

func buildWindowsClient() {
	// Build latest client version
	// TODO inject version
	log.Println("Building windows client")
	executable := "go"
	if runtime.GOOS == "linux" {
		executable = "/usr/local/go/bin/go"
	}
	cmd := exec.Command(
		executable,
		"build",
		"-o",
		"./downloads/tgl.exe",
		"./cmd/client/main.go")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if runtime.GOOS == "linux" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("GOPATH=%s", os.Getenv("GOPATH")))
		cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", os.Getenv("HOME")))
		cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s", os.Getenv("PATH")))
		cmd.Env = append(cmd.Env, "GOOS=windows")
		cmd.Env = append(cmd.Env, "GOARCH=386")
		cmd.Env = append(cmd.Env, "CGO_ENABLED=1")
		cmd.Env = append(cmd.Env, "CXX=i686-w64-mingw32-g++")
		cmd.Env = append(cmd.Env, "CC=i686-w64-mingw32-gcc")
	} else if runtime.GOOS == "darwin" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("GOPATH=%s", os.Getenv("GOPATH")))
		cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", os.Getenv("HOME")))
		cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s", os.Getenv("PATH")))
		cmd.Env = append(cmd.Env, "GOOS=windows")
		cmd.Env = append(cmd.Env, "GOARCH=amd64")
		cmd.Env = append(cmd.Env, "CGO_ENABLED=1")
		cmd.Env = append(cmd.Env, "CC=x86_64-w64-mingw32-gcc")
	}
	log.Println(cmd.Env)
	if err := cmd.Run(); err != nil {
		log.Println(out.String())
		log.Println(stderr.String())
		log.Println(err)
		return
	}
	log.Println("Completed building windows client")
}

func buildAddonZip() {
	// Build latest addon version
	// TODO inject version
	log.Println("Building zip addon files")
	buf := new(bytes.Buffer)

	w := zip.NewWriter(buf)

	var files = []string{
		"./TradeGuildLedger.iml",
		"./TradeGuildLedger.lua",
		"./TradeGuildLedger.txt",
	}
	for _, file := range files {
		f, err := w.Create(file)
		if err != nil {
			log.Println(err)
			return
		}
		content, err := ioutil.ReadFile(file)
		if err != nil {
			log.Println(err)
			return
		}
		_, err = f.Write(content)
		if err != nil {
			log.Println(err)
			return
		}
	}

	err := w.Close()
	if err != nil {
		log.Println(err)
		return
	}
	f, err := os.Create("./downloads/tgl.zip")
	if err != nil {
		log.Println(err)
		return
	}
	_, err = f.Write(buf.Bytes())
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Completed building zip addon files")
}
