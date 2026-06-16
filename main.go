package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configFile := flag.String("config", "config.json", "path to config JSON file")
	portFlag := flag.Int("port", 0, "port to listen on (overrides config)")
	flag.Parse()

	cfg, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	if *portFlag != 0 {
		cfg.Port = *portFlag
	}
	setConfig(cfg)

	// Reload config on SIGHUP
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGHUP)
		for range ch {
			newCfg, err := loadConfig(*configFile)
			if err != nil {
				log.Printf("reload failed: %v", err)
				continue
			}
			if *portFlag != 0 {
				newCfg.Port = *portFlag
			}
			setConfig(newCfg)
			log.Printf("config reloaded from %s", *configFile)
		}
	}()

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("testsvr listening on %s (config: %s)", addr, *configFile)
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
