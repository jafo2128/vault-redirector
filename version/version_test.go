package version

import (
  "testing"
  "fmt"
)

func TestDev(t *testing.T) {
  GitCommit = "mycommit"
  versionInfo := GetVersion()

  if versionInfo.Revision != "mycommit" {
    t.Fatalf("GetVersion has wrong Revision (got '%s' expected '%s')", versionInfo.Revision, "mycommit")
  }

  if versionInfo.Version != Version {
    t.Fatalf("GetVersion has wrong Version (got '%s' expected '%s')", versionInfo.Version, Version)
  }

  expectedVer := fmt.Sprintf("vault-redirector v%s-dev (mycommit)", Version)
  if versionInfo.String() != expectedVer {
    t.Fatalf("GetVersion String() has wrong value (got '%s' expected '%s')", versionInfo.String(), expectedVer)
  }
}

func TestRelease(t *testing.T) {
  GitCommit = "mycommit"
  GitDescribe = "1.2.3"
  versionInfo := GetVersion()

  if versionInfo.Revision != "mycommit" {
    t.Fatalf("GetVersion has wrong Revision (got '%s' expected '%s')", versionInfo.Revision, "mycommit")
  }

  if versionInfo.Version != "1.2.3" {
    t.Fatalf("GetVersion has wrong Version (got '%s' expected '%s')", versionInfo.Version, Version)
  }

  if versionInfo.VersionPrerelease != "" {
    t.Fatalf("GetVersion has wrong VersionPrerelease (got '%s' expected '%s')", versionInfo.VersionPrerelease, "dev")
  }

  expectedVer := fmt.Sprintf("vault-redirector v1.2.3")
  if versionInfo.String() != expectedVer {
    t.Fatalf("GetVersion String() has wrong value (got '%s' expected '%s')", versionInfo.String(), expectedVer)
  }
}
