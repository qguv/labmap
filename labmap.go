package main

import (
	"bytes"
	"fmt"
	"github.com/docopt/docopt.go"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const VERSION string = "0.1.0"
const USAGE string = "the LAB MAchine Probe v" + VERSION + `

Usage:
	labmap [options] <username>
	labmap -h | --help
	
Options:
	-k <keyfile>      location of private key
	-s <connections>  maximum simultaneous connections [default: 8]
	-c <command>      custom command to run
	-t <timeout>      time (s.) to wait for responses [default: 4]
	-p <placeholder>  what to print when command has no output
	-v, --version     display version
	-h, --help        display this screen

labmap helps you avoid the busiest CS machines. labmap maps a single command
across the W&M CS machines located on the firstfloor of MG-St Hall. By default,
labmap tells you how many users are (physically or remotely) logged in on each
system.

Examples:

	Displays least busy hosts first
		$ labmap you | awk '{print $2,$1}' | sort -n

	Displays busiest hosts first
		$ labmap you | awk '{print $2,$1}' | sort -nr

	Displays only free hosts
		$ labmap you | awk '{if (!$2) print $1}'

	Displays full uptime of each machine
		$ labmap -c 'uptime' you

	Displays names of connected users
		$ labmap -c 'users' -p '-----' you`

func cli_args() map[string]interface{} {
	m, err := docopt.Parse(
		USAGE,       // doc
		os.Args[1:], // argv
		true,        // help
		VERSION,     // version
		false,       // optionsFirst
	)
	if err != nil {
		return nil
	}
	return m
}

func async_each(max_connections, timeout int, keyfile, placeholder, username string, cmd []string) {

	subdomains := [...]string{
		"al", "anchor", "astro", "bart", "ca", "calvin", "daffy", "dilbert",
		"felix", "homer", "ia", "ickis", "il", "krumm", "me", "mn", "or", "pepe",
		"ren", "saranac", "steelhead", "stimpy", "tweety", "tx", "wi", "zippy",
	}

	c_print := make(chan string)
	c_conns := make(chan interface{}, max_connections)
	c_timeout := time.After(time.Duration(timeout) * time.Second)

	for _, sd := range subdomains {
		go func(sd string) {
			c_conns <- true
			c_print <- run_one(sd, keyfile, placeholder, username, cmd)
			<-c_conns
		}(sd)
	}

	for {
		select {
		case s := <-c_print:
			fmt.Print(s)
		case <-c_timeout:
			close(c_conns)
			close(c_print)
			return
		}
	}

}

func run_one(subdomain string, keyfile, placeholder, username string, command []string) (to_stdout string) {

	var precommand []string
	if keyfile != "" {
		precommand = append(precommand, "-i", keyfile)
	}

	precommand = append(
		precommand,
		"-o",
		"StrictHostKeyChecking=no",
		"-o",
		"UserKnownHostsFile=/dev/null",
		username+"@"+subdomain+".cs.wm.edu",
	)

	// instantiating new command struct
	cmd := exec.Command("ssh", append(precommand, command...)...)

	// creating return buffer and linking to command
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	// collecting output
	to_stdout += fmt.Sprintf("%-9v ", subdomain)
	if err != nil {
		to_stdout += fmt.Sprintln("failed!", err)
	}
	to_stdout += fmt.Sprint(out.String())

	// blit screen
	if !strings.HasSuffix(to_stdout, "\n") {
		to_stdout += placeholder + "\n"
	}

	return
}

func main() {

	// pull option flags from CLI args
	am := cli_args()
	command_flag := am["-c"]
	keyfile_flag := am["-k"]
	placeh_flag := am["-p"]

	// pull required and default arguments
	// these will never be nil
	username := am["<username>"].(string)
	max_conn, err := strconv.Atoi(am["-s"].(string))
	timeout, err := strconv.Atoi(am["-t"].(string))
	if err != nil {
		panic(err)
	}

	// get custom command or default to a user list
	var command []string
	if command_flag != nil {
		command = strings.Split(command_flag.(string), " ")
	} else {
		command = []string{"echo", "\"$(users | wc -w) users\""}
	}

	// get keyfile if available; default to ''
	var keyfile string
	if keyfile_flag != nil {
		keyfile = keyfile_flag.(string)
	}

	// get placeholder; default to ''
	var placeholder string
	if placeh_flag != nil {
		placeholder = placeh_flag.(string)
	}

	// run asynchronous process spawner
	async_each(max_conn, timeout, keyfile, placeholder, username, command)
}
