package main

import (
	"fmt"
	"flag"
	"github.com/golang/groupcache"
	"strings"
	"os"
	"os/signal"
	"syscall"
	"net/http"
	"crypto/md5"
)

var bind string
var peers string

func init() {
	flag.StringVar(&bind, "bind", "http://127.0.0.1:8080", "bind address like http://127.0.0.1:8080")
	flag.StringVar(&peers, "peers", "", "peers like, http://127.0.0.1:8080,http://127.0.0.1:8081,http://127.0.0.1:8082")
	flag.Parse()
}

func main() {
	peerSlice := strings.Split(peers, ",")
	if len(peerSlice) < 1 {
		fmt.Errorf("peers size is less than 1\n")
		os.Exit(1)
	}
	found := false
	for _, peer := range peerSlice {
		if peer == bind {
			found = true
			break
		}
	}
	if !found {
		fmt.Errorf("bind not in peers\n")
		os.Exit(1)
	}
	gcache := groupcache.NewHTTPPool(bind)
	gcache.Set(peerSlice...)

	md5cache := groupcache.NewGroup("md5cache", 64<<20, groupcache.GetterFunc(
        func(ctx groupcache.Context, key string, dest groupcache.Sink) error {
			fmt.Printf("calculate md5 sum for %q\n", key)
			m := md5.Sum([]byte(key))
			dest.SetBytes(m[0:])
			return nil
        }))

	handler := func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path[1:]
		var data []byte
		md5cache.Get(nil, name, groupcache.AllocatingByteSliceSink(&data))
		fmt.Fprintf(w, "%x", data)
	}

	http.HandleFunc("/", handler)

	go http.ListenAndServe(":8088", nil)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigs:
		fmt.Println("bye bye")
	}
}
