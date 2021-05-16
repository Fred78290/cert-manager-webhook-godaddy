package main

import (
	"os"
	"testing"

	"github.com/jetstack/cert-manager/test/acme/dns"
)

func TestRunsSuite(t *testing.T) {
	var zone string
	var manifest string
	var found bool

	if zone, found = os.LookupEnv("TEST_ZONE_NAME"); found == false {
		zone = "example.com"
	}

	if manifest, found = os.LookupEnv("TEST_MANIFEST_PATH"); found == false {
		manifest = "testdata/godaddy"
	}

	// The manifest path should contain a file named config.json that is a
	// snippet of valid configuration that should be included on the
	// ChallengeRequest passed as part of the test cases.

	fixture := dns.NewFixture(&godaddyDNSProviderSolver{},
		dns.SetResolvedZone(zone),
		dns.SetAllowAmbientCredentials(false),
		dns.SetManifestPath(manifest),
		dns.SetBinariesPath("__main__/hack/bin"),
	)

	fixture.RunConformance(t)
}
