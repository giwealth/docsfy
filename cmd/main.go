package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/giwealth/docsfy/internal/build"
	"github.com/giwealth/docsfy/internal/config"
	"github.com/giwealth/docsfy/internal/server"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		runBuild(os.Args[2:])
	case "serve":
		runServe(os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

func runBuild(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	docsDir := fs.String("docs", "./docs", "markdown docs directory")
	outDir := fs.String("out", "./dist", "output static directory")
	configPath := fs.String("config", "./docsfy.yaml", "config file path")
	siteTitle := fs.String("title", "", "site title override")
	krokiURL := fs.String("kroki-url", "", "Kroki server URL (overrides config; default https://kroki.io)")
	krokiDisable := fs.Bool("kroki-disable", false, "disable Kroki diagram rendering")
	_ = fs.Parse(args)

	cfg, err := config.Load(config.Options{
		ConfigPath:   *configPath,
		DocsDir:      *docsDir,
		OutDir:       *outDir,
		SiteTitle:    *siteTitle,
		KrokiURL:     *krokiURL,
		KrokiDisable: *krokiDisable,
	})
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	if err := build.Run(cfg); err != nil {
		log.Fatalf("build failed: %v", err)
	}

	fmt.Printf("build success: %s\n", cfg.OutDir)
}

func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	docsDir := fs.String("docs", "./docs", "markdown docs directory")
	port := fs.Int("port", 8080, "http port")
	configPath := fs.String("config", "./docsfy.yaml", "config file path")
	siteTitle := fs.String("title", "", "site title override")
	krokiURL := fs.String("kroki-url", "", "Kroki server URL (overrides config; default https://kroki.io)")
	krokiDisable := fs.Bool("kroki-disable", false, "disable Kroki diagram rendering")
	_ = fs.Parse(args)

	cfg, err := config.Load(config.Options{
		ConfigPath:   *configPath,
		DocsDir:      *docsDir,
		Port:         *port,
		SiteTitle:    *siteTitle,
		KrokiURL:     *krokiURL,
		KrokiDisable: *krokiDisable,
	})
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	if err := server.Run(cfg); err != nil {
		log.Fatalf("serve failed: %v", err)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  docsfy serve --docs ./docs --port 8080")
	fmt.Println("  docsfy build --docs ./docs --out ./dist")
}
