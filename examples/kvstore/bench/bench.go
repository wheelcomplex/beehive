// Benchmarks for the key value store.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const (
	keyLen     = 16
	minBodyLen = 16
	maxBodyLen = 512
)

var client *http.Client

type target struct {
	method string
	url    string
	body   []byte
}

func newTarget(method, url string, body []byte) target {
	return target{
		method: method,
		url:    url,
		body:   body,
	}
}

func (t *target) do() (d time.Duration, in int64, out int64, err error) {
	req, err := http.NewRequest(t.method, t.url, bytes.NewBuffer(t.body))
	if err != nil {
		return 0, 0, 0, err
	}
	in = req.ContentLength
	start := time.Now()
	res, err := client.Do(req)
	d = time.Since(start)
	if err != nil {
		return
	}
	defer func() {
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	out = int64(len(b))
	return
}

type result struct {
	Method string        `json:"method"`
	URL    string        `json:"url"`
	In     int64         `json:"in"`
	Out    int64         `json:"out"`
	Dur    time.Duration `json:"dur"`
	Err    string        `json:"err"`
}

func randRange(from, to int) int {
	return rand.Intn(to-from) + from
}

func randChar() byte {
	a := byte('a')
	z := byte('z')
	return byte(randRange(int(a), int(z)))
}

func randString(length int) string {
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		b[i] = randChar()
	}
	return string(b)
}

func generateTargets(addr string, writes, localReads,
	randReads int) []target {

	rand.Seed(time.Now().UTC().UnixNano())
	var targets []target
	var keys []string
	for i := 0; i < writes; i++ {
		k := randString(keyLen)
		keys = append(keys, k)
		bl := randRange(minBodyLen, maxBodyLen)
		t := newTarget("PUT", "http://"+addr+"/apps/kvstore/"+k,
			[]byte(randString(bl)))
		targets = append(targets, t)
	}
	for i := 0; i < localReads; i++ {
		t := newTarget("GET",
			"http://"+addr+"/apps/kvstore/"+keys[rand.Intn(len(keys))], []byte{})
		targets = append(targets, t)
	}
	for i := 0; i < randReads; i++ {
		t := newTarget("GET",
			"http://"+addr+"/apps/kvstore/"+randString(keyLen), []byte{})
		targets = append(targets, t)
	}
	return targets
}

func run(id int, targets []target, rounds int) []result {
	results := make([]result, 0, len(targets)*rounds)
	for i := 0; i < rounds; i++ {
		for j, t := range targets {
			fmt.Printf("%v-%v/%v-%s ", id, i, j, t.method)
			var err error
			res := result{
				Method: t.method,
				URL:    t.url,
			}
			res.Dur, res.In, res.Out, err = t.do()
			if err != nil {
				res.Err = err.Error()
			}
			results = append(results, res)
		}
	}
	return results
}

func mustSave(results []result, w io.Writer) {
	b, err := json.Marshal(results)
	if err != nil {
		panic(err)
	}
	w.Write(b)
}

var (
	addr    = flag.String("addr", "localhost:7767", "server address")
	writes  = flag.Int("writes", 10, "number of random keys to writes per round")
	localr  = flag.Int("localreads", 100, "number of reads from written keys")
	randr   = flag.Int("randomreads", 0, "number of random keys to read")
	rounds  = flag.Int("rounds", 1, "number of rounds")
	workers = flag.Int("workers", 1, "number of parallel clients")
	timeout = flag.Duration("timeout", 60*time.Second, "request timeout")
	output  = flag.String("out", "bench.out", "benchmark output file")
)

func main() {
	flag.Parse()

	f, err := os.Create(*output)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	client = &http.Client{Timeout: *timeout}

	ch := make(chan []result)
	for w := 0; w < *workers; w++ {
		go func(w int) {
			t := generateTargets(*addr, *writes, *localr, *randr)
			ch <- run(w, t, *rounds)
		}(w)
	}

	var res []result
	for i := 0; i < *workers; i++ {
		res = append(res, <-ch...)
	}
	mustSave(res, f)
}
