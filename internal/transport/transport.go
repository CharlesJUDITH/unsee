package transport

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/prometheus/common/log"
)

// TLSConfig provides a set of config options for configuring TLS connections
type TLSConfig struct {
	CAPath   string
	CertPath string
	KeyPath  string
}

// ReadJSON using one of supported transports (file:// http://)
func ReadJSON(uri string, timeout time.Duration, tlsConfig *tls.Config, target interface{}) error {
	u, err := url.Parse(uri)
	if err != nil {
		return err
	}
	var reader io.ReadCloser
	switch u.Scheme {
	case "http":
		reader, err = newHTTPReader(u.String(), timeout, nil)
	case "https":
		log.Infof("TLS Config: CA[%s] Cert[%s]", tlsConfig.RootCAs, tlsConfig.Certificates)
		reader, err = newHTTPReader(u.String(), timeout, &http.Transport{TLSClientConfig: tlsConfig})
	case "file":
		// if we have a file URI with relative path we need to expand it into an
		// absolute path, url.Parse doesn't support relative file paths
		if strings.HasPrefix(uri, "file:///") {
			reader, err = newFileReader(u.Path)
		} else {
			wd, e := os.Getwd()
			if e != nil {
				return e
			}
			absolutePath := path.Join(wd, strings.TrimPrefix(uri, "file://"))
			reader, err = newFileReader(absolutePath)
		}
	default:
		return fmt.Errorf("Unsupported URI scheme '%s' in '%s'", u.Scheme, u)
	}
	if err != nil {
		return err
	}
	defer reader.Close()
	return json.NewDecoder(reader).Decode(target)
}
