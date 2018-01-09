package config

import (
	"os"
	"testing"

	"github.com/pmezard/go-difflib/difflib"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// unset all unsee supported env variables before tests so we start with no
// config from previous test run
func resetEnv() {
	unseeEnvVariables := []string{
		"ALERTMANAGER_INTERVAL",
		"ALERTMANAGER_URI",
		"ANNOTATIONS_DEFAULT_HIDDEN",
		"ANNOTATIONS_HIDDEN",
		"ANNOTATIONS_VISIBLE",
		"CONFIG_DIR",
		"CONFIG_FILE",
		"DEBUG",
		"FILTERS_DEFAULT",
		"LABELS_COLOR_STATIC",
		"LABELS_COLOR_UNIQUE",
		"LABELS_KEEP",
		"LABELS_STRIP",
		"LISTEN_ADDRESS",
		"LISTEN_PORT",
		"LISTEN_PREFIX",
		"LOG_CONFIG",
		"LOG_LEVEL",
		"RECEIVERS_KEEP",
		"RECEIVERS_STRIP",
		"SENTRY_PRIVATE",
		"SENTRY_PUBLIC",

		"HOST",
		"PORT",
		"SENTRY_DSN",
	}
	for _, env := range unseeEnvVariables {
		os.Unsetenv(env)
	}
}

func testReadConfig(t *testing.T) {
	expectedConfig := `alertmanager:
  interval: 1s
  servers:
  - name: default
    uri: http://localhost
    timeout: 40s
    proxy: false
    tls:
      ca: ""
      cert: ""
      key: ""
annotations:
  default:
    hidden: true
  hidden: []
  visible:
  - summary
debug: true
filters:
  default:
  - '@state=active'
  - foo=bar
labels:
  keep:
  - foo
  - bar
  strip:
  - abc
  - def
  color:
    static:
    - a
    - bb
    - ccc
    unique:
    - f
    - gg
listen:
  address: 0.0.0.0
  port: 80
  prefix: /
log:
  config: true
  level: info
jira:
- regex: DEVOPS-[0-9]+
  uri: https://jira.example.com
- regex: FOO-[0-9]+
  uri: https://foo.example.com
receivers:
  keep: []
  strip: []
sentry:
  private: secret key
  public: public key
`

	configDump, err := yaml.Marshal(Config)
	if err != nil {
		t.Error(err)
	}

	if string(configDump) != expectedConfig {
		diff := difflib.UnifiedDiff{
			A:        difflib.SplitLines(expectedConfig),
			B:        difflib.SplitLines(string(configDump)),
			FromFile: "Expected",
			ToFile:   "Current",
			Context:  3,
		}
		text, err := difflib.GetUnifiedDiffString(diff)
		if err != nil {
			t.Error(err)
		}
		t.Errorf("Config mismatch:\n%s", text)
	}
}

func TestReadConfigLegacy(t *testing.T) {
	resetEnv()
	log.SetLevel(log.ErrorLevel)
	os.Setenv("ALERTMANAGER_TTL", "1s")
	os.Setenv("ALERTMANAGER_URIS", "default:http://localhost")
	os.Setenv("ANNOTATIONS_DEFAULT_HIDDEN", "true")
	os.Setenv("ANNOTATIONS_VISIBLE", "summary")
	os.Setenv("COLOR_LABELS_STATIC", "a bb ccc")
	os.Setenv("COLOR_LABELS_UNIQUE", "f gg")
	os.Setenv("DEBUG", "true")
	os.Setenv("FILTER_DEFAULT", "@state=active,foo=bar")
	os.Setenv("JIRA_REGEX", "DEVOPS-[0-9]+@https://jira.example.com FOO-[0-9]+@https://foo.example.com")
	os.Setenv("KEEP_LABELS", "foo bar")
	os.Setenv("STRIP_LABELS", "abc def")
	os.Setenv("SENTRY_DSN", "secret key")
	os.Setenv("SENTRY_PUBLIC_DSN", "public key")
	os.Setenv("HOST", "0.0.0.0")
	os.Setenv("PORT", "80")
	Config.Read()
	testReadConfig(t)
}

func TestReadConfig(t *testing.T) {
	resetEnv()
	log.SetLevel(log.ErrorLevel)
	os.Setenv("ALERTMANAGER_INTERVAL", "1s")
	os.Setenv("ALERTMANAGER_URI", "http://localhost")
	os.Setenv("ANNOTATIONS_DEFAULT_HIDDEN", "true")
	os.Setenv("ANNOTATIONS_VISIBLE", "summary")
	os.Setenv("DEBUG", "true")
	os.Setenv("FILTERS_DEFAULT", "@state=active foo=bar")
	os.Setenv("JIRA_REGEX", "DEVOPS-[0-9]+@https://jira.example.com FOO-[0-9]+@https://foo.example.com")
	os.Setenv("LABELS_COLOR_STATIC", "a bb ccc")
	os.Setenv("LABELS_COLOR_UNIQUE", "f gg")
	os.Setenv("LABELS_KEEP", "foo bar")
	os.Setenv("LABELS_STRIP", "abc def")
	os.Setenv("LISTEN_ADDRESS", "0.0.0.0")
	os.Setenv("LISTEN_PORT", "80")
	os.Setenv("SENTRY_PRIVATE", "secret key")
	os.Setenv("SENTRY_PUBLIC", "public key")
	Config.Read()
	testReadConfig(t)
}

type urlSecretTest struct {
	raw       string
	sanitized string
}

var urlSecretTests = []urlSecretTest{
	urlSecretTest{
		raw:       "http://localhost",
		sanitized: "http://localhost",
	},
	urlSecretTest{
		raw:       "http://alertmanager.example.com/path",
		sanitized: "http://alertmanager.example.com/path",
	},
	urlSecretTest{
		raw:       "http://user@alertmanager.example.com/path",
		sanitized: "http://user@alertmanager.example.com/path",
	},
	urlSecretTest{
		raw:       "https://user:password@alertmanager.example.com/path",
		sanitized: "https://user:xxx@alertmanager.example.com/path",
	},
	urlSecretTest{
		raw:       "file://localhost",
		sanitized: "file://localhost",
	},
}

func TestUrlSecretTest(t *testing.T) {
	for _, testCase := range urlSecretTests {
		sanitized := hideURLPassword(testCase.raw)
		if sanitized != testCase.sanitized {
			t.Errorf("Invalid sanitized url, expected '%s', got '%s'", testCase.sanitized, sanitized)
		}
	}
}
