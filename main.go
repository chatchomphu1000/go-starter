// Package main is the entry point for the go-starter application.
//
// @title           Go Starter API
// @version         1.0
// @description     Production-ready Go REST API starter with hexagonal architecture.
//
// @contact.name    API Support
//
// @host            localhost:8080
// @BasePath        /
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.
package main

import "github.com/chatchomphu1000/go-starter/cmd"

// version, commit, buildTime are injected via ldflags at build time.
var (
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

func main() {
	cmd.SetBuildInfo(version, commit, buildTime)
	cmd.Execute()
}
