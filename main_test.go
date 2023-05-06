package main

import (
	"os"
	"testing"

	"github.com/Fred78290/cert-manager-webhook-godaddy/test/acme/dns"
)

func TestRunsSuite(t *testing.T) {
	var zone string
	var manifest string
	var dnsServer string
	var found bool

	if zone, found = os.LookupEnv("TEST_ZONE_NAME"); found == false {
		zone = "example.com"
	}

	if zone, found = os.LookupEnv("TEST_DNS_SERVER"); found == false {
		dnsServer = "8.8.8.8:53"
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
		dns.SetDNSServer(dnsServer),
		dns.SetAllowAmbientCredentials(false),
		dns.SetManifestPath(manifest),
	)

	//fixture.RunConformance(t)
	fixture.RunBasic(t)
	fixture.RunExtended(t)
}
