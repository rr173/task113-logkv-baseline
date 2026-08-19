package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/chengjie/bytedance/logkv/internal/httpapi"
	"github.com/chengjie/bytedance/logkv/internal/kv"
)

func main() {
	smoke := flag.Bool("smoke-test", false, "run a self-contained smoke test and exit")
	dbPath := flag.String("db", "./logkv.db", "database file path")
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	if *smoke {
		if err := SmokeTest(); err != nil {
			fmt.Fprintln(os.Stderr, "smoke-test failed:", err)
			os.Exit(1)
		}
		fmt.Println("smoke-test OK")
		os.Exit(0)
	}

	store, err := kv.Open(*dbPath)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer store.Close()

	server := httpapi.NewServer(store)
	log.Printf("logkv listening on %s (db=%s)", *addr, *dbPath)
	if err := server.Start(*addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
