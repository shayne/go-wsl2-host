package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/shayne/go-wsl2-host/cmd/wsl2host/internal"
	"golang.org/x/sys/windows/svc"
)

func usage(errmsg string) {
	fmt.Fprintf(os.Stderr,
		"%s\n\n"+
			"usage: %s <command>\n"+
			"       where <command> is one of\n"+
			"       install, remove, debug, start, stop, pause or continue.\n",
		errmsg, os.Args[0])
	os.Exit(2)
}

func main() {
	const svcName = "wsl2host"

	isIntSess, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	}
	if !isIntSess {
		internal.RunService(svcName, false)
		return
	}

	if len(os.Args) < 2 {
		usage("no command specified")
	}

	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "debug":
		internal.RunService(svcName, true)
		return
	case "install":
		err = internal.InstallService(svcName, "WSL2 Host")
	case "remove":
		err = internal.RemoveService(svcName)
	case "start":
		err = internal.StartService(svcName)
	case "stop":
		err = internal.ControlService(svcName, svc.Stop, svc.Stopped)
	case "pause":
		err = internal.ControlService(svcName, svc.Pause, svc.Paused)
	case "continue":
		err = internal.ControlService(svcName, svc.Continue, svc.Running)
	default:
		usage(fmt.Sprintf("invalid command %s", cmd))
	}
	if err != nil {
		log.Fatalf("failed to %s %s: %v", cmd, svcName, err)
	}
	return
}
