package main

import (
	"fmt"
	. "mini-http/src"
	"os"
	"os/signal"
	"path"
	"syscall"
)

func main() {
	checkArgs()

	err := RunServer(os.Args[1:])

	if err != nil {
		fmt.Println("Mini HTTP Start Failed")
		os.Exit(1)
	}

	fmt.Println("Mini HTTP Started, Pressing CTRL + C to Shutdown")

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigChannel:
		fmt.Println("")
		fmt.Println("Mini HTTP Closed")
		os.Exit(0)
	}
}

func usage() {
	name := path.Base(os.Args[0])
	fmt.Printf("Usage of %s:\n", name)
	type flag struct {
		name         string
		description  string
		defaultValue string
		valueType    string
	}
	flags := []flag{
		{name: "port", description: "HTTP Port", defaultValue: "80", valueType: "int"},
		{name: "https-port", description: "HTTPS Port", defaultValue: "0", valueType: "int"},
		{name: "root", description: "WWW Root", defaultValue: "/www/", valueType: "string"},
		{name: "domain", description: "Domain", defaultValue: "", valueType: "string"},
		{name: "cert", description: "Domain Cert File", defaultValue: "", valueType: "string"},
		{name: "key", description: "Domain Key File", defaultValue: "", valueType: "string"},
		{name: "mode", description: "Set 'history' enable Single Page Routing", defaultValue: "", valueType: "string"},
		{name: "proxy", description: "Set proxy api", defaultValue: "", valueType: "string"},
		{name: "not-found", description: "Custom 404 page", defaultValue: "/404.html", valueType: "string"},
	}
	for i := 0; i < len(flags); i++ {
		f := flags[i]
		defaultValue := ""
		if f.defaultValue != "" {
			defaultValue = fmt.Sprintf("(default: %s)", f.defaultValue)
		}
		fmt.Printf(
			"    --%s  %s  %s\n%6s%s\n",
			fmt.Sprintf("%-10s", f.name),
			fmt.Sprintf("%-6s", f.valueType),
			defaultValue,
			"",
			f.description,
		)
	}
}

func checkArgs() {
	for i := 0; i < len(os.Args); i++ {
		if os.Args[i] == "-h" || os.Args[i] == "--help" || os.Args[i] == "-help" {
			usage()
			os.Exit(0)
		}
	}
}
