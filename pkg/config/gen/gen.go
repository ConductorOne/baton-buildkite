package main

import (
	cfg "github.com/conductorone/baton-buildkite/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("buildkite", cfg.Config)
}
