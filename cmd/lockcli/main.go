package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	lock "github.com/kramergroup-workflows/lock-agent/api"
	agent "github.com/kramergroup-workflows/lock-agent/internal"
)

var endpointPtr = flag.String("api-endpoint", os.Getenv("API_ENDPOINT"), "The API endpoint URL (defaults to $API_ENDPOINT env variable)")

func printUsage() {
	fmt.Println("lockcli [--api-endpoint api-url] {create workflow namespace | release id | delete id | get id}")
}

func main() {

	// Parse flags
	flag.Parse()

	if *endpointPtr == "" {
		log.Fatal("API endpoint not defined (set environment variable API_ENDPOINT or provide --api-endpoint flag)")
	}

	// Parse command
	if len(flag.Args()) < 2 {
		fmt.Println("ERROR: Insufficient number of arguments.")
		printUsage()
		os.Exit(1)
	}

	cmd := flag.Arg(0)

	api := agent.NewLockAPI(*endpointPtr)

	var err error

	switch cmd {

	case "create":
		workflow := flag.Arg(1)
		namespace := "default"
		if len(flag.Args()) > 2 {
			namespace = flag.Arg(2)
		}
		var lock lock.Lock
		lock, err = api.Create(workflow, namespace)
		fmt.Println(lock.ID)
		ioutil.WriteFile("/result", []byte(lock.ID), 0644)
		break

	case "get":
		id := flag.Arg(1)
		lock, err := api.Get(id)
		if err == nil {
			locJSON, _ := json.Marshal(lock)
			fmt.Print(string(locJSON))
			ioutil.WriteFile("/result", locJSON, 0644)
		}
		break

	case "release":
		id := flag.Arg(1)
		err = api.Release(id)
		break

	case "delete":
		id := flag.Arg(1)
		err = api.Delete(id)
		break

	default:
		fmt.Printf("Unknown command [%s]", flag.Arg(0))
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
