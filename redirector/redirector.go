package redirector

import (
    "errors"
    "fmt"
    "github.com/hashicorp/consul/api"
    "net/http"
    "os"
    "sync"
    "time"
)

// synchronize writes, to be safe.
var lock sync.Mutex

// Global variable to hold the active node host:port
var activeNodeHostPort string

// Global variable to determine whether to log or not
var enableLogging bool

// Global variable for Consul config
var ConsulConfig *api.Config

func Run(cconfig *api.Config, enable_log bool) {
  ConsulConfig = cconfig
  enableLogging = enable_log
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

// simple conditional logging
func log(line string) {
  if enableLogging {
    fmt.Printf("%s\n", line)
  }
}
