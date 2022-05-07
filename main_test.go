package main

import (
	"os"
	"testing"

	"github.com/cert-manager/cert-manager/test/acme/dns"
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
		dns.SetDNSName(zone),
		dns.SetDNSServer("10.0.0.5:53"),
		dns.SetAllowAmbientCredentials(false),
		dns.SetManifestPath(manifest),
	)

	//fixture.RunConformance(t)
	fixture.RunBasic(t)
	fixture.RunExtended(t)
}
