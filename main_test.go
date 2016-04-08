package main

import (
  "bytes"
  "os"
  "os/exec"
  "strings"
  "syscall"
  "testing"
)

// this methodology comes from https://talks.golang.org/2014/testing.slide#23
func TestUsage(t *testing.T) {
  // if this var is set, we're the exec'ed version, run the real (exiting) func
  if os.Getenv("BE_CRASHER") == "true" {
    usage()
    return
  }

  // buffers for command output and stderr, and waitStatus to get exit code
  cmdOutput := &bytes.Buffer{}
  cmdErr := &bytes.Buffer{}
  var waitStatus syscall.WaitStatus

  // build the command line to execute
  cmd := exec.Command(os.Args[0], "-test.run=TestUsage")
  cmd.Env = append(os.Environ(), "BE_CRASHER=true")
  cmd.Stdout = cmdOutput
  cmd.Stderr = cmdErr

  // run the command
  err := cmd.Run()
  // this SHOULD exit non-zero (1); fail if it doesn't
  if err == nil {
    t.Fatalf("should have exited 1, but exited 0")
  } else {
    // else err is not nil, so it (correctly) exited non-zero;
    // now check that exit code is 1
    if exitError, ok := err.(*exec.ExitError); ok {
      waitStatus = exitError.Sys().(syscall.WaitStatus)
      if waitStatus.ExitStatus() != 1 {
        t.Fatalf("should have exited 1, but exited %d", waitStatus.ExitStatus())
      }
    }
  }
  if string(cmdOutput.Bytes()) != "" {
    t.Fatalf("Expected no STDOUT but got: %s", string(cmdOutput.Bytes()))
  }
  if ! strings.Contains(string(cmdErr.Bytes()), usageMsg) {
    t.Fatalf("Expected usage message of '%s' in STDERR, but got '%s'", usageMsg, string(cmdErr.Bytes()))
  }
}
