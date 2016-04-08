package main

import (
  "os"
  "strings"
  "testing"
  "github.com/manheim/vault-redirector/helpers"
)

// this methodology comes from https://talks.golang.org/2014/testing.slide#23
func TestUsage(t *testing.T) {
  // if this var is set, we're the exec'ed version, run the real (exiting) func
  if os.Getenv("BE_CRASHER") == "true" {
    usage()
    return
  }

  cmdOutput, cmdError, exitStatus := helpers.RunCmdGetOutErrCode(os.Args[0], []string{"-test.run=TestUsage"})

  if exitStatus != 1 {
    t.Fatalf("Expected exit status 1 but got %d", exitStatus)
  }
  if cmdOutput != "" {
    t.Fatalf("Expected no STDOUT but got: %s", cmdOutput)
  }
  if ! strings.Contains(cmdError, usageMsg) {
    t.Fatalf("Expected usage message of '%s' in STDERR, but got '%s'", usageMsg, cmdError)
  }
}

// this methodology comes from https://talks.golang.org/2014/testing.slide#23
func TestConsulHostPortRegex(t *testing.T) {
  // if this var is set, we're the exec'ed version, run the real (exiting) func
  if os.Getenv("BE_CRASHER") == "true" {
    return
  }

  cmdOutput, cmdError, exitStatus := helpers.RunCmdGetOutErrCode(os.Args[0], []string{"-test.run=TestUsage"})

  if exitStatus != 1 {
    t.Fatalf("Expected exit status 1 but got %d", exitStatus)
  }
  if cmdOutput != "" {
    t.Fatalf("Expected no STDOUT but got: %s", cmdOutput)
  }
  if ! strings.Contains(cmdError, usageMsg) {
    t.Fatalf("Expected usage message of '%s' in STDERR, but got '%s'", usageMsg, cmdError)
  }
}
