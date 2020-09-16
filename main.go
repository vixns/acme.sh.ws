package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
)

// Cert struct
type Cert struct {
	Names          []string `json:"domains"`
	DNSAPI         string   `json:"dns_api"`
	ChallengeAlias string   `json:"challenge_alias"`
	KeyLength      string   `json:"key_length"`
}

func healthPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func validNames(names []string) bool {
	var r = regexp.MustCompile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9]|\*)\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)
	for _, n := range names {
		if r.MatchString(n) == false {
			return false
		}
	}
	return true
}

func issueCert(w http.ResponseWriter, r *http.Request) {
	c := Cert{}
	json.NewDecoder(r.Body).Decode(&c)
	if !validNames(c.Names) {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Invalid domain name(s) %v", c)
		return
	}
	var acmesh string
	acmesh, ok := os.LookupEnv("ACME_SH_PATH")
	if !ok {
		acmesh = "/usr/local/bin/acme.sh"
	}
	if c.KeyLength == "" {
		c.KeyLength = "4096"
	}
	cmdargs := []string{acmesh, "--issue", "--keylength", c.KeyLength}
	if os.Getenv("DRY_RUN") != "" {
		cmdargs = append(cmdargs, "--test")
	}
	if c.DNSAPI == "" {
		cmdargs = append(cmdargs, "-w", os.Getenv("WEBROOT_DIR"))
	} else {
		cmdargs = append(cmdargs, "--dns", "dns_"+c.DNSAPI, "--dnssleep", "60")
		if c.ChallengeAlias != "" {
			cmdargs = append(cmdargs, "--challenge-alias", c.ChallengeAlias)
		}
	}
	for _, n := range c.Names {
		cmdargs = append(cmdargs, "-d", n)
	}
	cmd := exec.Command(acmesh)
	cmd.Args = cmdargs
	out, err := cmd.CombinedOutput()
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "error: %s %v", out, cmdargs)
		log.Printf("error: %s %v\n", out, cmdargs)
		return
	}
	if os.Getenv("DEPLOY_HOOK") != "" {
		name := c.Names[0]
		deployargs := []string{acmesh, "--deploy", "--deploy-hook", os.Getenv("DEPLOY_HOOK"), "-d", name}
		if strings.HasPrefix(c.KeyLength, "ec-") {
			deployargs = append(deployargs, "--ecc")
		}

		cmd = exec.Command(acmesh)
		cmd.Args = deployargs
		out, err = cmd.CombinedOutput()
	}
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "error: %s", out)
		log.Printf("error: %s\n", out)
		return
	}
	fmt.Fprintf(w, "Success: %s", out)
}

func deleteCert(w http.ResponseWriter, r *http.Request) {
	c := Cert{}
	json.NewDecoder(r.Body).Decode(&c)
	if !validNames(c.Names) {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Invalid domain name(s) %v", c)
		return
	}
	var acmesh string
	acmesh, ok := os.LookupEnv("ACME_SH_PATH")
	if !ok {
		acmesh = "/usr/local/bin/acme.sh"
	}
	cmdargs := []string{acmesh, "--remove"}
	cmd := exec.Command(acmesh)
	for _, n := range c.Names {
		cmdargs = append(cmdargs, "-d", n)
	}
	cmd.Args = cmdargs
	out, err := cmd.CombinedOutput()
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "error: %s", out)
		log.Printf("error: %s\n", out)
		return
	}
	fmt.Fprintf(w, "OK")
}

func renewCerts(w http.ResponseWriter, r *http.Request) {
	var acmesh string
	acmesh, ok := os.LookupEnv("ACME_SH_PATH")
	if !ok {
		acmesh = "/usr/local/bin/acme.sh"
	}
	cmdargs := []string{acmesh, "--cron", "-w", os.Getenv("WEBROOT_DIR")}
	cmd := exec.Command(acmesh)
	cmd.Args = cmdargs
	out, err := cmd.CombinedOutput()
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "error: %s", out)
		log.Printf("error: %s\n", out)
		return
	}
	log.Printf("Success: %s\n", out)
	fmt.Fprintf(w, "Success: %s", out)
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", healthPage).Methods("GET")
	myRouter.HandleFunc("/", issueCert).Methods("POST")
	myRouter.HandleFunc("/", deleteCert).Methods("DELETE")
	myRouter.HandleFunc("/renew", renewCerts).Methods("GET")
	var bindIP string
	bindIP, ok := os.LookupEnv("BIND_IP")
	if !ok {
		bindIP = "0.0.0.0"
	}
	var bindPort string
	bindPort, bpok := os.LookupEnv("BIND_PORT")
	if !bpok {
		bindPort = "3000"
	}
	log.Fatal(http.ListenAndServe(bindIP+":"+bindPort, myRouter))
}

func main() {
	handleRequests()
}
