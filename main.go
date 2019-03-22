package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type Cert struct {
	Names          []string `json:"domains"`
	DnsApi         string   `json:"dns_api"`
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
	acmesh := "/usr/local/bin/acme.sh"
	if os.Getenv("ACME_SH_PATH") != "" {
		acmesh = os.Getenv("ACME_SH_PATH")
	}
	if c.KeyLength == "" {
		c.KeyLength = "4096"
	}
	cmdargs := []string{acmesh, "--issue", "--keylength", c.KeyLength}
	if os.Getenv("DRY_RUN") != "" {
		cmdargs = append(cmdargs, "--test")
	}
	if c.DnsApi == "" {
		cmdargs = append(cmdargs, "-w", os.Getenv("WEBROOT_DIR"))
	} else {
		cmdargs = append(cmdargs, "--dns", "dns_"+c.DnsApi, "--dnssleep", "60")
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
	acmesh := "/usr/local/bin/acme.sh"
	if os.Getenv("ACME_SH_PATH") != "" {
		acmesh = os.Getenv("ACME_SH_PATH")
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

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", healthPage).Methods("GET")
	myRouter.HandleFunc("/", issueCert).Methods("POST")
	myRouter.HandleFunc("/", deleteCert).Methods("DELETE")
	bind_ip := "0.0.0.0"
	if os.Getenv("BIND_IP") != "" {
		bind_ip = os.Getenv("BIND_IP")
	}
	bind_port := "3000"
	if os.Getenv("BIND_PORT") != "" {
		bind_port = os.Getenv("BIND_PORT")
	}
	log.Fatal(http.ListenAndServe(bind_ip+":"+bind_port, myRouter))
}

func main() {
	handleRequests()
}
