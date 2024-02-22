package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	addrFlag       = flag.String("addr", ":80", "address to listen on")
	secretFlag     = flag.String("secret", "./secret", "path to secret file")
	masterHookFlag = flag.String("master-hook", "./master-hook", "path to executable to run for push to master branch")
	tagHookFlag    = flag.String("tag-hook", "./tag-hook", "path to executable to run for new tags")
)

type Payload struct {
	Ref     string `json:"ref"`
	RefType string `json:"ref_type"`
}

func main() {
	flag.Parse()

	secretBytes, err := os.ReadFile(*secretFlag)
	if err != nil {
		log.Fatalln("failed to read secret file:", err)
	}
	secret := bytes.TrimRight(secretBytes, "\n")

	masterCh := make(chan struct{}, 1)
	tagCh := make(chan string, 32)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("handling request")
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("failed to read HTTP request:", err)
			return
		}
		digest := r.Header.Get("X-Hub-Signature")
		ok := checkMAC(secret, payload, digest)
		log.Println("MAC check result:", ok)
		if !ok {
			return
		}
		parsed := &Payload{}
		if err := json.Unmarshal(payload, parsed); err != nil {
			log.Println("failed to parse payload:", err)
			return
		}

		switch {
		case parsed.RefType == "tag":
			tag := parsed.Ref
			select {
			case tagCh <- tag:
			default:
			}
		case parsed.Ref == "refs/heads/master":
			select {
			case masterCh <- struct{}{}:
			default:
			}
		default:
			log.Println("not master push or tag creation, ignoring payload", parsed)
		}
	})

	execCh := make(chan []string)

	go func() {
		for range masterCh {
			execCh <- []string{*masterHookFlag}
		}
	}()

	go func() {
		for tag := range tagCh {
			execCh <- []string{*tagHookFlag, tag}
		}
	}()

	go func() {
		for cmd := range execCh {
			execHook(cmd)
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

func execHook(cmd []string) {
	log.Println("going to run hook", cmd)
	p, err := os.StartProcess(cmd[0], cmd,
		&os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
	if err != nil {
		log.Println("error in StartProcess:", err)
		return
	}
	state, err := p.Wait()
	if err != nil {
		log.Println("error in Wait:", err)
		return
	}
	log.Println("hook exit status:", state)
}
