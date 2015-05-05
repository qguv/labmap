package main

import (
	"bytes"
	"fmt"
	"github.com/docopt/docopt.go"
	"github.com/howeyc/gopass"
	"golang.org/x/crypto/ssh"
	"os"
	"time"
)

const VERSION string = "1.0.0"
const USAGE string = "the LAB MAchine Probe v" + VERSION + `

Usage:
	labmap [options] <username>
	labmap -h | --help
	
Options:
	-p <password>     override prompt and use this password
	-c <command>      custom command to run
	-d <downtime>     time to wait between connections [default: 0s]
	-t <timeout>      time to wait for responses [default: 5s]
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
		$ labmap -c 'users' --placeholder '-----' you`

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

func async_each(timeout, downtime time.Duration, username, password, cmd string) {

	subdomains := [...]string{
		"al", "anchor", "astro", "bart", "ca", "calvin", "daffy", "dilbert",
		"felix", "homer", "ia", "ickis", "il", "krumm", "me", "mn", "or", "pepe",
		"ren", "saranac", "steelhead", "stimpy", "tweety", "tx", "wi", "zippy",
	}

<<<<<<< HEAD
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
=======
	ret := make(chan string)
	for _, sd := range subdomains {
		fmt.Println("Running on", sd)
		go run_command(sd, username, password, cmd, ret)
		time.Sleep(downtime)
	}

	abort := time.After(timeout)
	for range subdomains {
		select {
		case m := <-ret:
			fmt.Printf(m)
		case <-abort:
			fmt.Println("timed out")
			return
		}
	}
}

func run_command(hostname, username, password, command string, ret chan string) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}
	client, err := ssh.Dial("tcp", hostname+".cs.wm.edu:22", config)
	if err != nil {
		ret <- "Failed to dial: " + err.Error()
		return
	}

	session, err := client.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(command); err != nil {
		panic("Failed to run: " + err.Error())
	}

	ret <- hostname + ": " + b.String()
}

func main() {
	opts := cli_args()
	username := opts["<username>"].(string)

	timeout, err := time.ParseDuration(opts["-t"].(string))
	if err != nil {
		panic(err)
	}

	downtime, err := time.ParseDuration(opts["-d"].(string))
	if err != nil {
		panic(err)
	}

	var command string
	if opts["-c"] != nil {
		command = opts["-c"].(string)
	} else {
		command = "echo \"$(users | wc -w) users\""
	}

	var password string
	if opts["-p"] != nil {
		password = opts["-p"].(string)
	} else {
		fmt.Printf("%s's Password: ", username)
		password = string(gopass.GetPasswdMasked())
	}

	async_each(timeout, downtime, username, password, command)
>>>>>>> overhaul with ssh library
}
