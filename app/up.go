package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/hex"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	addrFlag   = flag.String("addr", ":80", "address to listen on")
	secretFlag = flag.String("secret", "./secret", "path to secret file")
	hookFlag   = flag.String("hook", "./hook", "path to executable to run")
)

func main() {
	flag.Parse()

	secretBytes, err := ioutil.ReadFile(*secretFlag)
	if err != nil {
		log.Fatalln("failed to read secret file:", err)
	}
	secret := bytes.TrimRight(secretBytes, "\n")
	_ = secret

	reqCh := make(chan bool, 1000)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("handling request")
		payload, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("failed to read HTTP request:", err)
			return
		}
		digest := r.Header.Get("X-Hub-Signature")
		req := checkMAC(secret, payload, digest)
		log.Println("MAC check result:", req)
		select {
		case reqCh <- req:
		default:
		}
	})

	go func() {
		for req := range reqCh {
			if !req {
				log.Println("received null request")
				continue
			}
			log.Println("going to run hook", *hookFlag)
			p, err := os.StartProcess(*hookFlag, []string{*hookFlag},
				&os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
			if err != nil {
				log.Println("error in StartProcess:", err)
				continue
			}
			state, err := p.Wait()
			if err != nil {
				log.Println("error in Wait:", err)
				continue
			}
			log.Println("hook exit status:", state)
		}
	}()

	log.Println("going to listen", *addrFlag)
	log.Fatal(http.ListenAndServe(*addrFlag, nil))
}

func checkMAC(secret, payload []byte, digest string) bool {
	mac := hmac.New(sha1.New, secret)
	mac.Write(payload)
	expectedDigest := "sha1=" + hex.EncodeToString(mac.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(expectedDigest), []byte(digest)) == 1
}
