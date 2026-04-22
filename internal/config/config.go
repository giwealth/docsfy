package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DocsDir       string `yaml:"docsDir"`
	OutDir        string `yaml:"outDir"`
	Port          int    `yaml:"port"`
	SiteTitle     string `yaml:"siteTitle"`
	AssetsDir     string `yaml:"assetsDir"`
	TemplatesDir  string `yaml:"templatesDir"`
	KrokiURL      string `yaml:"krokiUrl"`
	KrokiDisabled bool   `yaml:"krokiDisabled"`
	FrontendMermaid bool `yaml:"frontendMermaid"`
}

type Options struct {
	ConfigPath   string
	DocsDir      string
	OutDir       string
	Port         int
	SiteTitle    string
	KrokiURL     string
	KrokiDisable bool
}

func Load(opts Options) (Config, error) {
	cfg := Config{
		DocsDir:       "./docs",
		OutDir:        "./dist",
		Port:          8080,
		SiteTitle:     "Docsfy",
		AssetsDir:     "./web/assets",
		TemplatesDir:  "./web/templates",
		KrokiURL:      "https://kroki.io",
		KrokiDisabled: false,
		FrontendMermaid: true,
	}

	if opts.ConfigPath != "" {
		if b, err := os.ReadFile(opts.ConfigPath); err == nil {
			_ = yaml.Unmarshal(b, &cfg)
		}
	}

	if opts.DocsDir != "" {
		cfg.DocsDir = opts.DocsDir
	}
	if opts.OutDir != "" {
		cfg.OutDir = opts.OutDir
	}
	if opts.Port != 0 {
		cfg.Port = opts.Port
	}
	if opts.SiteTitle != "" {
		cfg.SiteTitle = opts.SiteTitle
	}
	if opts.KrokiURL != "" {
		cfg.KrokiURL = opts.KrokiURL
	}
	if opts.KrokiDisable {
		cfg.KrokiDisabled = true
	}

	if cfg.DocsDir == "" {
		return Config{}, errors.New("docs directory is required")
	}
	if cfg.TemplatesDir == "" || cfg.AssetsDir == "" {
		return Config{}, errors.New("template/assets directory is required")
	}
	return cfg, nil
}
