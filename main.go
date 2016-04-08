package main

import (
    "errors"
    "flag"
    "fmt"
    "github.com/hashicorp/consul/api"
    "net/http"
    "os"
    "regexp"
    "sync"
    "time"
    "github.com/manheim/vault-redirector/version"
)

// Global variable for Consul config
var ConsulConfig *api.Config

// Global variable to determine whether to log or not
var enableLogging bool

// Global variable to hold the active node host:port
var activeNodeHostPort string

// synchronize writes, to be safe.
var lock sync.Mutex

// Global constant for CLI usage
const usageMsg string = "goredirector [-verbose] CONSUL_HOST:PORT\n"

func handler(w http.ResponseWriter, r *http.Request) {
  path := r.URL.Path[1:]
  log(fmt.Sprintf("REQUEST for: %s", path))
  // format the redirect URL
  redir_to := fmt.Sprintf("https://%s/%s", activeNodeHostPort, path)
  log(fmt.Sprintf("REDIRECT 307 to: %s", redir_to))
  // redirect
  http.Redirect(w, r, redir_to, 307)
}

// should be called in a goroutine with GOMAXPROCS >= 2
func pollConsul() {
  // loop infinitely; update every 5 seconds
  for {
    updateActiveFromConsul()
    time.Sleep(5 * time.Second)
  }
}

// should ONLY be called from within pollConsul
func updateActiveFromConsul() {
  // set an initial active node, to make sure we don't error before handling requests
  active, err := getActiveFromConsul()
  if err != nil {
    fmt.Printf("Error polling from Consul: %s", err.Error())
  } else {
    lock.Lock()
    defer lock.Unlock()
    activeNodeHostPort = active
  }
}

func getActiveFromConsul() (string, error) {
  log("Querying Consul")

  // Consul API client using our global config
  client, err := api.NewClient(ConsulConfig)
  if err != nil {
    return "", err
  }

  // Health check API endpoint
  health := client.Health()

  // get the service we're interested in, only passing checks
  svc_entry, _, err := health.Service("vault", "", true, nil)
  if err != nil {
    return "", err
  }

  // iterate the service entries with passing checks
  for _, svc := range svc_entry {
    // iterate each check
    for _, check := range svc.Checks {
      if (check.CheckID == "service:vault") && (check.Status == "passing") {
        node_port := fmt.Sprintf("%s:%d", check.Node, svc.Service.Port)
        log(fmt.Sprintf("Found passing service:vault check: %s", node_port))
        return node_port, nil
      }
    }
  }

  return "", errors.New("no passing service found")
}

func usage() {
  fmt.Printf(usageMsg)
  flag.PrintDefaults()
  os.Exit(1)
}

// simple conditional logging
func log(line string) {
  if enableLogging {
    fmt.Printf("%s\n", line)
  }
}

func main() {
  showVersion := false
  // command line flag for enabling logging (slows down responses)
  flag.BoolVar(&enableLogging, "verbose", false, "Run with log output")
  // command line flag to show version
  flag.BoolVar(&showVersion, "version", false, "Show version information and exit")
  // add usage
  flag.Usage = usage
  // parse command line
  flag.Parse()

  // show version and exit, if called with -version
  if showVersion {
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
  log(fmt.Sprintf("Connecting to Consul at %s", consulHostPort))
  ConsulConfig = api.DefaultNonPooledConfig()
  ConsulConfig.Address = consulHostPort
  ConsulConfig.Scheme = "http"

  // set an initial active node, to make sure we don't error before handling requests
  active, err := getActiveFromConsul()
  if err != nil {
    log("Got error on initial request from Consul; exiting")
    log(err.Error())
    os.Exit(1)
  }
  activeNodeHostPort = active

  // run the background polling goroutine
  log("Starting Consul polling goroutine")
  go pollConsul()

  http.HandleFunc("/", handler)
  log("Starting server listening on port 8080")
  http.ListenAndServe(":8080", nil)
}
