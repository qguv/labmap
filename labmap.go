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
	-t <timeout>      time (s.) to wait for responses [default: 5]
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

	for _, sd := range subdomains {
		go run_one(sd, keyfile, placeholder, username, cmd)
	}

	time.Sleep(time.Duration(timeout) * time.Second)
}

func run_one(subdomain string, keyfile, placeholder, username string, command []string) {

	var precommand []string
	if keyfile != "" {
		precommand = append(precommand, "-i", keyfile)
	}
	precommand = append(precommand, username+"@"+subdomain+".cs.wm.edu")

	// instantiating new command struct
	cmd := exec.Command("ssh", append(precommand, command...)...)

	// creating return buffer and linking to command
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	// collecting output
	var s string
	s += fmt.Sprintf("%-9v ", subdomain)
	if err != nil {
		s += fmt.Sprintln("failed!", err)
	}
	s += fmt.Sprint(out.String())

	// blit screen
	if !strings.HasSuffix(s, "\n") {
		s += placeholder + "\n"
	}
	fmt.Print(s)
}

func main() {
	am := cli_args()

	var command []string
	if am["-c"] != nil {
		command = strings.Split(am["-c"].(string), " ")
	} else {
		command = []string{"echo", "\"$(users | wc -w) users\""}
	}

	var keyfile string
	if am["-k"] != nil {
		keyfile = am["-k"].(string)
	}

	username := am["<username>"].(string)

	var placeholder string
	if am["-p"] != nil {
		placeholder = am["-p"].(string)
	}

	max_conn, err := strconv.Atoi(am["-s"].(string))
	timeout, err := strconv.Atoi(am["-t"].(string))
	if err != nil {
		panic(err)
	}

	async_each(max_conn, timeout, keyfile, placeholder, username, command)
}
