package dex

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/argoproj/argo-cd/v2/common"
	"github.com/argoproj/argo-cd/v2/util/errors"
)

func decorateDirector(director func(req *http.Request), target *url.URL) func(req *http.Request) {
	return func(req *http.Request) {
		director(req)
		req.Host = target.Host
	}
}

type DexTLSConfig struct {
	DisableTLS       bool
	StrictValidation bool
	RootCAs          *x509.CertPool
	Certificate      []byte
}

func TLSConfig(tlsConfig *DexTLSConfig) *tls.Config {
	if tlsConfig == nil || tlsConfig.DisableTLS {
		return nil
	}
	/* False positive: All network communication is performed over TLS including service-to-service communication between the three components (argocd-server, argocd-repo-server, argocd-application-controller).

	Argo CD provides two inbound TLS endpoints that can be configured: (1) the user-facing endpoint of the argocd-server workload which serves the UI and the API and (2) the endpoint of the argocd-repo-server, 
	which is accessed by argocd-server and argocd-application-controller workloads to request repository operations. By default, and without further configuration, both of these endpoints will be set-up to use 
	an automatically generated, self-signed certificate. */

	/* #nosec G402 */
	if !tlsConfig.StrictValidation {
		return &tls.Config{ 
			InsecureSkipVerify: true, // #nosec G402
			MinVersion: tls.VersionTLS12,
		}
	}
	return &tls.Config{
		InsecureSkipVerify: false,
		RootCAs:            tlsConfig.RootCAs,
		MinVersion: tls.VersionTLS12,
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			if !bytes.Equal(rawCerts[0], tlsConfig.Certificate) {
				return fmt.Errorf("dex server certificate does not match")
			}
			return nil
		},
	}
}

// NewDexHTTPReverseProxy returns a reverse proxy to the Dex server. Dex is assumed to be configured
// with the external issuer URL muxed to the same path configured in server.go. In other words, if
// Argo CD API server wants to proxy requests at /api/dex, then the dex config yaml issuer URL should
// also be /api/dex (e.g. issuer: https://argocd.example.com/api/dex)
func NewDexHTTPReverseProxy(serverAddr string, baseHRef string, tlsConfig *DexTLSConfig) func(writer http.ResponseWriter, request *http.Request) {

	fullAddr := DexServerAddressWithProtocol(serverAddr, tlsConfig)

	target, err := url.Parse(fullAddr)
	errors.CheckError(err)
	target.Path = baseHRef

	proxy := httputil.NewSingleHostReverseProxy(target)

	if tlsConfig != nil && !tlsConfig.DisableTLS {
		proxy.Transport = &http.Transport{
			TLSClientConfig: TLSConfig(tlsConfig),
		}
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
		if resp.StatusCode == 500 {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			err = resp.Body.Close()
			if err != nil {
				return err
			}
			log.WithFields(log.Fields{
				common.SecurityField: common.SecurityMedium,
			}).Errorf("received error from dex: %s", string(b))
			resp.ContentLength = 0
			resp.Header.Set("Content-Length", strconv.Itoa(0))
			resp.Header.Set("Location", fmt.Sprintf("%s?has_sso_error=true", path.Join(baseHRef, "login")))
			resp.StatusCode = http.StatusSeeOther
			resp.Body = io.NopCloser(bytes.NewReader(make([]byte, 0)))
			return nil
		}
		return nil
	}
	proxy.Director = decorateDirector(proxy.Director, target)
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}

// NewDexRewriteURLRoundTripper creates a new DexRewriteURLRoundTripper
func NewDexRewriteURLRoundTripper(dexServerAddr string, T http.RoundTripper) DexRewriteURLRoundTripper {
	dexURL, _ := url.Parse(dexServerAddr)
	return DexRewriteURLRoundTripper{
		DexURL: dexURL,
		T:      T,
	}
}

// DexRewriteURLRoundTripper is an HTTP RoundTripper to rewrite HTTP requests to the specified
// dex server address. This is used when reverse proxying Dex to avoid the API server from
// unnecessarily communicating to Argo CD through its externally facing load balancer, which is not
// always permitted in firewalled/air-gapped networks.
type DexRewriteURLRoundTripper struct {
	DexURL *url.URL
	T      http.RoundTripper
}

func (s DexRewriteURLRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r.URL.Host = s.DexURL.Host
	r.URL.Scheme = s.DexURL.Scheme
	return s.T.RoundTrip(r)
}

func DexServerAddressWithProtocol(orig string, tlsConfig *DexTLSConfig) string {
	if strings.Contains(orig, "://") {
		return orig
	} else {
		if tlsConfig == nil || tlsConfig.DisableTLS {
			return "http://" + orig
		} else {
			return "https://" + orig
		}
	}
}
