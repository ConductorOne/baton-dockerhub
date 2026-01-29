package main

import (
	cfg "github.com/conductorone/baton-dockerhub/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("dockerhub", cfg.Configuration)
}
