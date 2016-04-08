package main

import (
    "flag"
    "fmt"
    "github.com/hashicorp/consul/api"
    "os"
    "regexp"
    "github.com/manheim/vault-redirector/version"
    "github.com/manheim/vault-redirector/redirector"
)

// Global constant for CLI usage
const usageMsg string = "vault-redirector [-verbose] CONSUL_HOST:PORT\n"

func usage() {
  fmt.Fprintf(os.Stderr, usageMsg)
  flag.PrintDefaults()
  os.Exit(1)
}

func main() {
  // command line flag for enabling logging (slows down responses)
  enableLogging := flag.Bool("verbose", false, "Run with log output")
  // command line flag to show version
  showVersion := flag.Bool("version", false, "Show version information and exit")

  flag.Usage = usage
  flag.Parse()

  // show version and exit, if called with -version
  if *showVersion {
    versionInfo := version.GetVersion()
    println(versionInfo.String())
    os.Exit(0)
  }

  // if positional argument was omitted, die
  if len(flag.Args()) == 0 {
    usage()
  }

  consulHostPort := flag.Args()[0]

  // check that it appears to be (more or less) the right format
  matched, _ := regexp.MatchString("^.*:[0-9]+$", consulHostPort)
  if ! matched {
    fmt.Fprintf(os.Stderr, "ERROR: CONSUL_HOST_PORT (%s) does not match /^.*:[0-9]+$/", consulHostPort)
    os.Exit(1)
  }

  // Consul API client config
  //log(fmt.Sprintf("Connecting to Consul at %s", consulHostPort))
  ConsulConfig := api.DefaultNonPooledConfig()
  ConsulConfig.Address = consulHostPort
  ConsulConfig.Scheme = "http"

  redirector.Run(ConsulConfig, *enableLogging)
}
