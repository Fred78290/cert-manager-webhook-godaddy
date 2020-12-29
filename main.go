package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/klog/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/jetstack/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/cmd"
	"github.com/jetstack/cert-manager/pkg/issuer/acme/dns/util"

	pkgutil "github.com/jetstack/cert-manager/pkg/util"
)

var phVersion = "v0.0.0-unset"
var phBuildDate = ""

// Context wrapper
type Context struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// NewContext return a context. Timeout is in seconds
func NewContext(timeout time.Duration) Context {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	return Context{
		ctx:    ctx,
		cancel: cancel,
	}
}

// DNSRecord a DNS record
type DNSRecord struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Data     string `json:"data"`
	Priority int    `json:"priority,omitempty"`
	TTL      int    `json:"ttl,omitempty"`
}

// GroupName a API group name
var GroupName = os.Getenv("GROUP_NAME")

func main() {

	versionPtr := flag.Bool("version", false, "Give the version")

	// Declare it
	flag.String("tls-cert-file", "/tls/tls.crt", "tls-cert-file")
	flag.String("tls-private-key-file", "/tls/tls.key", "tls-private-key-file")

	flag.Parse()

	if *versionPtr {
		klog.Infof("The current version is:%s, build at:%s", phVersion, phBuildDate)
	} else {

		if GroupName == "" {
			panic("GROUP_NAME must be specified")
		}

		klog.Infof("Launch cert-manager-webhook-godaddy with group name: %s", GroupName)

		// This will register our godaddy DNS provider with the webhook serving
		// library, making it available as an API under the provided GroupName.
		// You can register multiple DNS provider implementations with a single
		// webhook, where the Name() method will be used to disambiguate between
		// the different implementations.
		cmd.RunWebhookServer(GroupName,
			&godaddyDNSProviderSolver{},
		)
	}
}

// godaddyDNSProviderSolver implements the provider-specific logic needed to
// 'present' an ACME challenge TXT record for your own DNS provider.
// To do so, it must implement the `github.com/jetstack/cert-manager/pkg/acme/webhook.Solver`
// interface.
type godaddyDNSProviderSolver struct {
	// If a Kubernetes 'clientset' is needed, you must:
	// 1. uncomment the additional `client` field in this structure below
	// 2. uncomment the "k8s.io/client-go/kubernetes" import at the top of the file
	// 3. uncomment the relevant code in the Initialize method below
	// 4. ensure your webhook's service account has the required RBAC role
	//    assigned to it for interacting with the Kubernetes APIs you need.
	client *kubernetes.Clientset
}

// LocalObjectReference A reference to an object in the same namespace as the referent.
// If the referent is a cluster-scoped resource (e.g. a ClusterIssuer),
// the reference instead refers to the resource with the given name in the
// configured 'cluster resource namespace', which is set as a flag on the
// controller component (and defaults to the namespace that cert-manager
// runs in).
type LocalObjectReference struct {
	// Name of the resource being referred to.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	Name *string `json:"name,omitempty"`
}

// SecretKeySelector A reference to a specific 'key' within a Secret resource.
// In some instances, `key` is a required field.
type SecretKeySelector struct {
	// The name of the Secret resource being referred to.
	LocalObjectReference `json:",inline"`

	// The key of the entry in the Secret resource's `data` field to be used.
	// Some instances of this field may be defaulted, in others it may be
	// required.
	// +optional
	Key    string `json:"key,omitempty"`
	Secret string `json:"secret,omitempty"`
}

// godaddyDNSProviderConfig is a structure that is used to decode into when
// solving a DNS01 challenge.
// This information is provided by cert-manager, and may be a reference to
// additional configuration that's needed to solve the challenge for this
// particular certificate or issuer.
// This typically includes references to Secret resources containing DNS
// provider credentials, in cases where a 'multi-tenant' DNS solver is being
// created.
// If you do *not* require per-issuer or per-certificate configuration to be
// provided to your webhook, you can skip decoding altogether in favour of
// using CLI flags or similar to provide configuration.
// You should not include sensitive information here. If credentials need to
// be used by your provider here, you should reference a Kubernetes Secret
// resource and fetch these credentials using a Kubernetes clientset.
type godaddyDNSProviderConfig struct {
	// These fields will be set by users in the
	// `issuer.spec.acme.dns01.providers.webhook.config` field.

	APIKeySecretRef SecretKeySelector `json:"apiKeySecret"`
	Production      bool              `json:"production"`
	TTL             int               `json:"ttl"`
}

// Name is used as the name for this DNS solver when referencing it on the ACME
// Issuer resource.
// This should be unique **within the group name**, i.e. you can have two
// solvers configured with the same Name() **so long as they do not co-exist
// within a single webhook deployment**.
// For example, `cloudflare` may be used as the name of a solver.
func (c *godaddyDNSProviderSolver) Name() string {
	return "godaddy"
}

// Present is responsible for actually presenting the DNS record with the
// DNS provider.
// This method should tolerate being called multiple times with the same value.
// cert-manager itself will later perform a self check to ensure that the
// solver has correctly configured the DNS provider.
func (c *godaddyDNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return err
	}

	klog.V(6).Infof("Decoded configuration %v", cfg)

	recordName := c.extractRecordName(ch.ResolvedFQDN, ch.ResolvedZone)

	dnsZone, err := c.getZone(ch.ResolvedZone)
	if err != nil {
		return err
	}

	rec := []DNSRecord{
		{
			Type: "TXT",
			Name: recordName,
			Data: ch.Key,
			TTL:  cfg.TTL,
		},
	}

	klog.Infof("Present record: %s with key: %s", recordName, ch.Key)

	return c.updateRecords(cfg, ch.ResourceNamespace, rec, dnsZone, recordName)
}

// CleanUp should delete the relevant TXT record from the DNS provider console.
// If multiple TXT records exist with the same record name (e.g.
// _acme-challenge.example.com) then **only** the record with the same `key`
// value provided on the ChallengeRequest should be cleaned up.
// This is in order to facilitate multiple DNS validations for the same domain
// concurrently.
func (c *godaddyDNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return err
	}

	klog.V(5).Infof("Decoded configuration %v", cfg)

	recordName := c.extractRecordName(ch.ResolvedFQDN, ch.ResolvedZone)

	dnsZone, err := c.getZone(ch.ResolvedZone)
	if err != nil {
		klog.Errorf("Unable to get zone:%s, error: %v", ch.ResolvedZone, err)

		return err
	}

	rec := []DNSRecord{
		{
			Type: "TXT",
			Name: recordName,
			Data: "null",
		},
	}

	klog.Infof("Cleanup record: %s", recordName)

	return c.updateRecords(cfg, ch.ResourceNamespace, rec, dnsZone, recordName)
}

// Initialize will be called when the webhook first starts.
// This method can be used to instantiate the webhook, i.e. initialising
// connections or warming up caches.
// Typically, the kubeClientConfig parameter is used to build a Kubernetes
// client that can be used to fetch resources from the Kubernetes API, e.g.
// Secret resources containing credentials used to authenticate with DNS
// provider accounts.
// The stopCh can be used to handle early termination of the webhook, in cases
// where a SIGTERM or similar signal is sent to the webhook process.
func (c *godaddyDNSProviderSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}

	c.client = cl

	return nil
}

// loadConfig is a small helper function that decodes JSON configuration into
// the typed config struct.
func loadConfig(cfgJSON *extapi.JSON) (godaddyDNSProviderConfig, error) {
	cfg := godaddyDNSProviderConfig{}
	// handle the 'base case' where no configuration has been provided
	if cfgJSON == nil {
		klog.Errorln("Config is not defined")

		return cfg, nil
	}

	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		klog.Errorf("Can't decode config: %v", err)

		return cfg, fmt.Errorf("error decoding solver config: %v", err)
	}

	return cfg, nil
}

func (c *godaddyDNSProviderSolver) updateRecords(cfg godaddyDNSProviderConfig, resourceNamespace string, records []DNSRecord, domainZone string, recordName string) error {
	body, err := json.Marshal(records)
	if err != nil {
		return err
	}

	authAPIKey, authAPISecret, e := c.getAPIKey(cfg, resourceNamespace)
	if e != nil {
		return e
	}

	// https://developer.godaddy.com/doc/endpoint/domains
	// OTE environment: https://api.ote-godaddy.com
	// PRODUCTION environment: https://api.godaddy.com
	baseURL := "https://api.ote-godaddy.com"
	if cfg.Production {
		baseURL = "https://api.godaddy.com"
	}

	var resp *http.Response
	url := fmt.Sprintf("/v1/domains/%s/records/TXT/%s", domainZone, recordName)
	resp, err = c.makeRequest(cfg, *authAPIKey, *authAPISecret, baseURL, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		klog.Errorf("Unable to request: %s%s, got error:%s", baseURL, url, err)

		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		errStr := fmt.Sprintf("could not create record %v; Status: %v; Body: %s", string(body), resp.StatusCode, string(bodyBytes))

		klog.Errorln(errStr)

		return fmt.Errorf(errStr)
	}

	return nil
}

func (c *godaddyDNSProviderSolver) makeRequest(cfg godaddyDNSProviderConfig, authAPIKey string, authAPISecret string, baseURL string, method string, uri string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", baseURL, uri), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", pkgutil.CertManagerUserAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", authAPIKey, authAPISecret))

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	return client.Do(req)
}

func (c *godaddyDNSProviderSolver) extractRecordName(fqdn, domain string) string {
	if idx := strings.Index(fqdn, "."+domain); idx != -1 {
		return fqdn[:idx]
	}

	return util.UnFqdn(fqdn)
}

func (c *godaddyDNSProviderSolver) extractDomainName(zone string) string {
	authZone, err := util.FindZoneByFqdn(zone, util.RecursiveNameservers)
	if err != nil {
		return zone
	}

	return util.UnFqdn(authZone)
}

func (c *godaddyDNSProviderSolver) getZone(fqdn string) (string, error) {
	authZone, err := util.FindZoneByFqdn(fqdn, util.RecursiveNameservers)
	if err != nil {
		return "", err
	}

	return util.UnFqdn(authZone), nil
}

func (c *godaddyDNSProviderSolver) getAPIKey(cfg godaddyDNSProviderConfig, namespace string) (*string, *string, error) {

	ctx := NewContext(time.Minute * 2)
	defer ctx.cancel()

	if cfg.APIKeySecretRef.LocalObjectReference.Name != nil {
		secretName := *cfg.APIKeySecretRef.LocalObjectReference.Name

		klog.V(6).Infof("try to load secret `%s` in namespace:`%s`", secretName, namespace)

		sec, err := c.client.CoreV1().Secrets(namespace).Get(ctx.ctx, secretName, metav1.GetOptions{})
		if err != nil {
			return nil, nil, fmt.Errorf("unable to get secret `%s`; %v", secretName, err)
		}

		keyBytes, ok := sec.Data[cfg.APIKeySecretRef.Key]
		if !ok {
			return nil, nil, fmt.Errorf("key %s not found in secret \"%s/%s\"", cfg.APIKeySecretRef.Key, secretName, namespace)
		}

		secretBytes, ok := sec.Data[cfg.APIKeySecretRef.Secret]
		if !ok {
			return nil, nil, fmt.Errorf("secret %s not found in secret \"%s/%s\"", cfg.APIKeySecretRef.Secret, secretName, namespace)
		}

		apiKey := string(keyBytes)
		apiSecret := string(secretBytes)
		return &apiKey, &apiSecret, nil
	}

	klog.V(6).Infof("GoDaddy use key pair %s:%s", cfg.APIKeySecretRef.Key, cfg.APIKeySecretRef.Secret)

	return &cfg.APIKeySecretRef.Key, &cfg.APIKeySecretRef.Secret, nil
}
