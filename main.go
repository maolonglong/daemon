package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

var (
	out   = flag.String("out", "/dev/null", "redirect output to ...")
	child = flag.Bool("child", false, "!!! internal use only !!!")
)

func help() {
	fmt.Fprintf(
		os.Stderr,
		"Usage of %[1]v:\n  %[1]v [options] [--] <command> [args]\n\n",
		os.Args[0],
	)
	fmt.Fprintf(os.Stderr, "Example:\n  %v python3 -m http.server\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %v -out daemon.log -- python3 -m http.server\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
}

func fork1() error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(self, append([]string{"-child"}, os.Args[1:]...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	return cmd.Run()
}

func fork2() error {
	cmd := exec.Command(flag.Arg(0), flag.Args()[1:]...)
	if *out != "/dev/null" {
		f, err := os.OpenFile(*out, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
		if err != nil {
			return err
		}
		defer f.Close()
		cmd.Stdout = f
		cmd.Stderr = f
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	fmt.Println(cmd.Process.Pid)
	return nil
}

func main() {
	flag.Usage = help
	flag.Parse()

	if flag.NArg() == 0 {
		help()
		os.Exit(2)
	}

	var err error
	if !*child {
		err = fork1()
	} else {
		err = fork2()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "daemonize failed: %v\n", err)
		os.Exit(1)
	}
}
