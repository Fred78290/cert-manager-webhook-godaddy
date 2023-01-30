package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/cert-manager/cert-manager/pkg/issuer/acme/dns/util"

	cmdutil "github.com/cert-manager/cert-manager/cmd/util"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd/server"

	logf "github.com/cert-manager/cert-manager/pkg/logs"
	pkgutil "github.com/cert-manager/cert-manager/pkg/util"
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
	Type     string  `json:"type"`
	Name     string  `json:"name"`
	Data     string  `json:"data"`
	TTL      int     `json:"ttl,omitempty"`
	Priority *int    `json:"priority,omitempty"`
	Weight   *int    `json:"weight,omitempty"`
	Protocol *string `json:"protocol,omitempty"`
	Service  *string `json:"service,omitempty"`
}

// GroupName a API group name
var GroupName = os.Getenv("GROUP_NAME")

func runWebhookServer(groupName string, hooks ...webhook.Solver) {
	stopCh, exit := cmdutil.SetupExitHandler(cmdutil.GracefulShutdown)
	defer exit() // This function might call os.Exit, so defer last

	logs.InitLogs()
	defer logs.FlushLogs()

	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	cmd := server.NewCommandStartWebhookServer(os.Stdout, os.Stderr, stopCh, groupName, hooks...)
	//cmd.Version = fmt.Sprintf("The current version is:%s, build at:%s", phVersion, phBuildDate)

	cmd.Flags().AddGoFlagSet(flag.CommandLine)

	klog.Infof("Launch cert-manager-webhook-godaddy with group name: %s", GroupName)

	if err := cmd.Execute(); err != nil {
		logf.Log.Error(err, "error executing command")
		cmdutil.SetExitCode(err)
	}
}

func main() {
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}

	// This will register our godaddy DNS provider with the webhook serving
	// library, making it available as an API under the provided GroupName.
	// You can register multiple DNS provider implementations with a single
	// webhook, where the Name() method will be used to disambiguate between
	// the different implementations.
	runWebhookServer(GroupName,
		&godaddyDNSProviderSolver{},
	)
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

	APIKeySecretRef SecretKeySelector `json:"apiKeySecretRef"`
	Production      bool              `json:"production"`
	TTL             int               `json:"ttl"`
}

func (c godaddyDNSProviderConfig) goDaddyURL() string {
	// https://developer.godaddy.com/doc/endpoint/domains
	// OTE environment: https://api.ote-godaddy.com
	// PRODUCTION environment: https://api.godaddy.com
	if c.Production {
		return "https://api.godaddy.com"
	}

	return "https://api.ote-godaddy.com"
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

	klog.V(4).Infof("Decoded configuration %v", cfg)

	recordName := c.extractRecordName(ch.ResolvedFQDN, ch.ResolvedZone)

	dnsZone, err := c.getZone(ch.ResolvedFQDN)
	if err != nil {
		klog.Errorf("Unable to get zone:%s, error: %v", ch.ResolvedZone, err)
		return err
	}

	rec := DNSRecord{
		Type: "TXT",
		Name: recordName,
		Data: ch.Key,
		TTL:  cfg.TTL,
	}

	klog.Infof("Present record: %s on zone: %s with key: %s", recordName, dnsZone, ch.Key)

	return c.addRecord(cfg, ch.ResourceNamespace, rec, dnsZone, recordName)
}

func (c *godaddyDNSProviderSolver) getAllRecords(authAPIKey, authAPISecret, baseURL, domainZone string) ([]DNSRecord, error) {
	var records []DNSRecord

	url := fmt.Sprintf("/v1/domains/%s/records", domainZone)
	resp, err := c.makeRequest(authAPIKey, authAPISecret, baseURL, http.MethodGet, url, nil)
	if err != nil {
		klog.Errorf("Unable to request: %s%s, got error:%s", baseURL, url, err)

		return nil, err
	}

	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		errStr := fmt.Sprintf("Unable to list records for zone: %s; Status: %v; Body: %s", domainZone, resp.StatusCode, string(bodyBytes))

		klog.Errorln(errStr)

		return nil, fmt.Errorf(errStr)
	}

	if err := json.Unmarshal(bodyBytes, &records); err != nil {
		klog.Errorf("Can't decode config: %v", err)

		return nil, fmt.Errorf("error decoding solver config: %v", err)
	}

	return records, nil
}

func (c *godaddyDNSProviderSolver) deleteRecord(authAPIKey, authAPISecret, baseURL, domainZone string, record *DNSRecord) error {
	var body []byte

	url := fmt.Sprintf("/v1/domains/%s/records/%s/%s", domainZone, record.Type, record.Name)

	resp, err := c.makeRequest(authAPIKey, authAPISecret, baseURL, http.MethodDelete, url, bytes.NewReader(body))
	if err != nil {
		klog.Errorf("Unable to request: %s%s, got error:%s", baseURL, url, err)

		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		errStr := fmt.Sprintf("Unable to delete records for zone: %s; Status: %v; Body: %s", domainZone, resp.StatusCode, string(bodyBytes))

		klog.Errorln(errStr)

		return fmt.Errorf(errStr)
	}

	return nil
}

// CleanUp should delete the relevant TXT record from the DNS provider console.
// If multiple TXT records exist with the same record name (e.g.
// _acme-challenge.example.com) then **only** the record with the same `key`
// value provided on the ChallengeRequest should be cleaned up.
// This is in order to facilitate multiple DNS validations for the same domain
// concurrently.
func (c *godaddyDNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	var records []DNSRecord

	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return err
	}

	authAPIKey, authAPISecret, err := c.getAPIKey(cfg, ch.ResourceNamespace)
	if err != nil {
		return err
	}

	baseURL := cfg.goDaddyURL()

	klog.V(4).Infof("Decoded configuration %v", cfg)

	recordName := c.extractRecordName(ch.ResolvedFQDN, ch.ResolvedZone)

	dnsZone, err := c.getZone(ch.ResolvedFQDN)
	if err != nil {
		klog.Errorf("Unable to get zone:%s, error: %v", ch.ResolvedZone, err)
		return err
	}

	klog.Infof("Cleanup record: %s on zone: %s with key: %s", recordName, dnsZone, ch.Key)

	if records, err = c.getAllRecords(*authAPIKey, *authAPISecret, baseURL, dnsZone); err != nil {
		klog.Errorf("Unable to fetch records from zone:%s, error: %v", dnsZone, err)
		return err
	}

	for _, record := range records {
		if record.Name == recordName && record.Data == ch.Key && record.Type == "TXT" {
			if err = c.deleteRecord(*authAPIKey, *authAPISecret, baseURL, dnsZone, &record); err == nil {
				klog.Infof("Cleaned record: %s on zone: %s with key: %s", recordName, dnsZone, ch.Key)
			} else {
				klog.ErrorS(err, "Cleaned record: %s on zone: %s with key: %s", recordName, dnsZone, ch.Key)
			}
			return err
		}
	}

	klog.Warningf("Record %s is not found in zone %s", recordName, dnsZone)

	return nil
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

func (c *godaddyDNSProviderSolver) addRecord(cfg godaddyDNSProviderConfig, resourceNamespace string, record DNSRecord, domainZone string, recordName string) error {
	body, err := json.Marshal([]DNSRecord{record})
	if err != nil {
		return err
	}

	authAPIKey, authAPISecret, e := c.getAPIKey(cfg, resourceNamespace)
	if e != nil {
		return e
	}

	baseURL := cfg.goDaddyURL()

	var resp *http.Response
	url := fmt.Sprintf("/v1/domains/%s/records/%s/%s", domainZone, record.Type, recordName)
	resp, err = c.makeRequest(*authAPIKey, *authAPISecret, baseURL, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		klog.Errorf("Unable to request: %s%s, got error:%s", baseURL, url, err)

		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		errStr := fmt.Sprintf("could not create record %v; Status: %v; Body: %s", string(body), resp.StatusCode, string(bodyBytes))

		klog.Errorln(errStr)

		return fmt.Errorf(errStr)
	}

	return nil
}

func (c *godaddyDNSProviderSolver) makeRequest(authAPIKey string, authAPISecret string, baseURL string, method string, uri string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", baseURL, uri), body)
	if err != nil {
		return nil, err
	}

	userAgent := fmt.Sprintf("%s/%s (%s) cert-manager/%s",
		"godaddy-webhook",
		phVersion, pkgutil.VersionInfo().Platform, phBuildDate)

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
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

func (c *godaddyDNSProviderSolver) getZone(fqdn string) (string, error) {
	authZone, err := util.FindZoneByFqdn(fqdn, util.RecursiveNameservers)
	if err != nil {
		return "", err
	}

	return util.UnFqdn(authZone), nil
}

func (c *godaddyDNSProviderSolver) getAPIKey(cfg godaddyDNSProviderConfig, namespace string) (*string, *string, error) {

	ctx := NewContext(120)
	defer ctx.cancel()

	if cfg.APIKeySecretRef.LocalObjectReference.Name != nil {
		secretName := *cfg.APIKeySecretRef.LocalObjectReference.Name

		klog.V(4).Infof("try to load secret `%s` in namespace:`%s`", secretName, namespace)

		sec, err := c.client.CoreV1().Secrets(namespace).Get(ctx.ctx, secretName, metav1.GetOptions{})
		if err != nil {
			klog.V(4).ErrorS(err, "unable to get secret `%s`", secretName)
			return nil, nil, fmt.Errorf("unable to get secret `%s`; %v", secretName, err)
		}

		klog.V(4).Infof("Secret `%s` in namespace:`%s` found", secretName, namespace)

		keyBytes, ok := sec.Data[cfg.APIKeySecretRef.Key]
		if !ok {
			klog.V(4).Info("key %s not found in secret \"%s/%s\"", cfg.APIKeySecretRef.Key, secretName, namespace)
			return nil, nil, fmt.Errorf("key %s not found in secret \"%s/%s\"", cfg.APIKeySecretRef.Key, secretName, namespace)
		}

		secretBytes, ok := sec.Data[cfg.APIKeySecretRef.Secret]
		if !ok {
			klog.V(4).Info("secret %s not found in secret \"%s/%s\"", cfg.APIKeySecretRef.Secret, secretName, namespace)
			return nil, nil, fmt.Errorf("secret %s not found in secret \"%s/%s\"", cfg.APIKeySecretRef.Secret, secretName, namespace)
		}

		apiKey := string(keyBytes)
		apiSecret := string(secretBytes)

		klog.V(4).Infof("GoDaddy use key pair %s:%s", apiKey, apiSecret)

		return &apiKey, &apiSecret, nil
	}

	klog.V(4).Infof("GoDaddy use key pair %s:%s", cfg.APIKeySecretRef.Key, cfg.APIKeySecretRef.Secret)

	return &cfg.APIKeySecretRef.Key, &cfg.APIKeySecretRef.Secret, nil
}
