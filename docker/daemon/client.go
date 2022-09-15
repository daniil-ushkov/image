package daemon

import (
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/daniil-ushkov/image/v5/types"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/go-connections/tlsconfig"
)

const (
	// The default API version to be used in case none is explicitly specified
	defaultAPIVersion = "1.22"
)

// NewDockerClient initializes a new API client based on the passed SystemContext.
func newDockerClient(sys *types.SystemContext) (*dockerclient.Client, error) {
	host := dockerclient.DefaultDockerHost
	if sys != nil && sys.DockerDaemonHost != "" {
		host = sys.DockerDaemonHost
	}

	// Sadly, unix:// sockets don't work transparently with dockerclient.NewClient.
	// They work fine with a nil httpClient; with a non-nil httpClient, the transport’s
	// TLSClientConfig must be nil (or the client will try using HTTPS over the PF_UNIX socket
	// regardless of the values in the *tls.Config), and we would have to call sockets.ConfigureTransport.
	//
	// We don't really want to configure anything for unix:// sockets, so just pass a nil *http.Client.
	//
	// Similarly, if we want to communicate over plain HTTP on a TCP socket, we also need to set
	// TLSClientConfig to nil. This can be achieved by using the form `http://`
	url, err := dockerclient.ParseHostURL(host)
	if err != nil {
		return nil, err
	}
	var httpClient *http.Client
	if url.Scheme != "unix" {
		if url.Scheme == "http" {
			hc, err := httpConfig(sys)
			if err != nil {
				return nil, err
			}
			httpClient = hc
		} else {
			hc, err := tlsConfig(sys)
			if err != nil {
				return nil, err
			}
			httpClient = hc
		}
	}

	return dockerclient.NewClient(host, defaultAPIVersion, httpClient, nil)
}

func tlsConfig(sys *types.SystemContext) (*http.Client, error) {
	options := tlsconfig.Options{}
	if sys != nil && sys.DockerDaemonInsecureSkipTLSVerify {
		options.InsecureSkipVerify = true
	}

	if sys != nil && sys.DockerDaemonCertPath != "" {
		options.CAFile = filepath.Join(sys.DockerDaemonCertPath, "ca.pem")
		options.CertFile = filepath.Join(sys.DockerDaemonCertPath, "cert.pem")
		options.KeyFile = filepath.Join(sys.DockerDaemonCertPath, "key.pem")
	}

	tlsc, err := tlsconfig.Client(options)
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{
		TLSClientConfig: tlsc,
	}

	if sys.Proxy != "" {
		proxyUrl, err := url.Parse(sys.Proxy)
		if err != nil {
			return nil, err
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
	}

	return &http.Client{
		Transport:     tr,
		CheckRedirect: dockerclient.CheckRedirect,
	}, nil
}

func httpConfig(sys *types.SystemContext) (*http.Client, error) {
	tr := &http.Transport{
		TLSClientConfig: nil,
	}

	if sys.Proxy != "" {
		proxyUrl, err := url.Parse(sys.Proxy)
		if err != nil {
			return nil, nil
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
	}

	client := &http.Client{
		Transport:     tr,
		CheckRedirect: dockerclient.CheckRedirect,
	}

	return client, nil
}
