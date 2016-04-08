package helpers

import (
  "bytes"
  "os"
  "os/exec"
  "syscall"
)

// run a command; return its STDOUT, STDERR and Exit Status (code)
// this is ONLY used by tests!
func RunCmdGetOutErrCode(cmd_str string, args []string) (string, string, int) {
  // buffers for command output and stderr, and waitStatus to get exit code
  cmdOutput := &bytes.Buffer{}
  cmdErr := &bytes.Buffer{}
  var waitStatus syscall.WaitStatus
  exitStatus := -1

  cmd := exec.Command(cmd_str, args...)
  //cmd := exec.Command(os.Args[0], "-test.run=TestUsage")
  cmd.Env = append(os.Environ(), "BE_CRASHER=true")
  cmd.Stdout = cmdOutput
  cmd.Stderr = cmdErr

  // run the command
  err := cmd.Run()
  // this SHOULD exit non-zero (1); fail if it doesn't
  if err == nil {
    exitStatus = 0
  } else {
    // else err is not nil, so it exited non-zero; get the exit code
    if exitError, ok := err.(*exec.ExitError); ok {
      waitStatus = exitError.Sys().(syscall.WaitStatus)
      exitStatus = waitStatus.ExitStatus()
    }
  }

  return string(cmdOutput.Bytes()), string(cmdErr.Bytes()), exitStatus
}
