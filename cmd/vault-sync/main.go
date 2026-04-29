package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/example/vault-sync/internal/config"
	vsync "github.com/example/vault-sync/internal/sync"
)

var version = "dev"

func main() {
	var (
		cfgFile    = flag.String("config", ".vault-sync.yaml", "path to config file")
		showVer    = flag.Bool("version", false, "print version and exit")
		dryRun     = flag.Bool("dry-run", false, "print secrets count without writing")
	)
	flag.Parse()

	if *showVer {
		fmt.Printf("vault-sync %s\n", version)
		os.Exit(0)
	}

	log.SetFlags(log.Ltime | log.Lshortfile)

	cfg, err := config.Load(*cfgFile)
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	if *dryRun {
		log.Printf("[dry-run] vault=%s path=%s output=%s prefixes=%v",
			cfg.VaultAddr, cfg.SecretPath, cfg.OutputFile, cfg.Prefixes)
		os.Exit(0)
	}

	s, err := vsync.New(cfg)
	if err != nil {
		log.Fatalf("initialising syncer: %v", err)
	}

	res, err := s.Run()
	if err != nil {
		log.Fatalf("sync failed: %v", err)
	}

	fmt.Printf("✓ synced %d/%d secret(s) → %s\n",
		res.SecretsWritten, res.SecretsTotal, res.OutputFile)
}
