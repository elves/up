package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io/ioutil"
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

	secretBytes, err := ioutil.ReadFile(*secretFlag)
	if err != nil {
		log.Fatalln("failed to read secret file:", err)
	}
	secret := bytes.TrimRight(secretBytes, "\n")
	_ = secret

	masterCh := make(chan struct{}, 1)
	tagCh := make(chan string, 32)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("handling request")
		payload, err := ioutil.ReadAll(r.Body)
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
		req := Req{}
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

	go func() {
		for range masterCh {
			execHook(*masterHookFlag)
		}
	}()

	go func() {
		for tag := range tagCh {
			execHook(*tagHookFlag, tag)
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

func execHook(hook string, args ...string) {
	log.Println("going to run hook", hook, args)
	p, err := os.StartProcess(hook, append([]string{hook}, args...),
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
