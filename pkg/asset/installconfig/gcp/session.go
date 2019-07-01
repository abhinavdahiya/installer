package gcp

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	googleoauth "golang.org/x/oauth2/google"
	"gopkg.in/AlecAivazis/survey.v1"
)

var (
	authEnvs            = []string{"GOOGLE_CREDENTIALS", "GOOGLE_CLOUD_KEYFILE_JSON", "GCLOUD_KEYFILE_JSON"}
	defaultAuthFilePath = filepath.Join(os.Getenv("HOME"), ".gcp", "osServiceAccount.json")
)

// Session is an object representing session for GCP API.
type Session struct {
	Credentials *googleoauth.Credentials
}

// GetSession returns an GCP session by using credentials found in default locations in order:
// env GOOGLE_CREDENTIALS,
// env GOOGLE_CLOUD_KEYFILE_JSON,
// env GCLOUD_KEYFILE_JSON,
// file ~/.gcp/osServiceAccount.json, and
// gcloud cli defaults
// and, if no creds are found, asks for them and stores them on disk in a config file
func GetSession(ctx context.Context) (*Session, error) {
	creds, err := loadCredentials(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load credentials")
	}

	return &Session{
		Credentials: creds,
	}, nil
}

func loadCredentials(ctx context.Context) (*googleoauth.Credentials, error) {
	var loaders []credLoader
	for _, env := range authEnvs {
		loaders = append(loaders, &envLoader{env: env})
	}
	loaders = append(loaders, &fileLoader{path: defaultAuthFilePath})
	loaders = append(loaders, &cliLoader{})

	for _, l := range loaders {
		creds, err := l.Load(ctx)
		if err != nil {
			logrus.Debug(errors.Wrapf(err, "failed to load credentials from %s", l))
			continue
		}
		return creds, nil
	}
	return getCredentials(ctx)
}

func getCredentials(ctx context.Context) (*googleoauth.Credentials, error) {
	creds, err := (&cliLoader{}).Load(ctx)
	if err != nil {
		return nil, err
	}

	filePath := defaultAuthFilePath
	logrus.Info("Saving the credentials to %q", filePath)
	if err := os.MkdirAll(filepath.Dir(filePath), 0700); err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(filePath, creds.JSON, 0600); err != nil {
		return nil, err
	}
	return creds, nil
}

type credLoader interface {
	Load(context.Context) (*googleoauth.Credentials, error)
}

type envLoader struct {
	env string
}

func (e *envLoader) Load(ctx context.Context) (*googleoauth.Credentials, error) {
	var content string
	if f := os.Getenv(e.env); len(f) > 0 {
		content = f
	}

	return (&fileOrContentLoader{content: content}).Load(ctx)
}

func (e *envLoader) String() string {
	return fmt.Sprintf("loading from environment variable %q", e.env)
}

type fileOrContentLoader struct {
	content string
}

func (fc *fileOrContentLoader) Load(ctx context.Context) (*googleoauth.Credentials, error) {
	// if this is a path and we can stat it, assume it's ok
	if _, err := os.Stat(fc.content); err == nil {
		return (&fileLoader{path: fc.content}).Load(ctx)
	}

	return googleoauth.CredentialsFromJSON(ctx, []byte(fc.content))
}

func (fc *fileOrContentLoader) String() string {
	return fmt.Sprintf("loading from file or content %q", fc.content)
}

type fileLoader struct {
	path string
}

func (f *fileLoader) Load(ctx context.Context) (*googleoauth.Credentials, error) {
	content, err := ioutil.ReadFile(f.path)
	if err != nil {
		return nil, err
	}
	return googleoauth.CredentialsFromJSON(ctx, []byte(content))
}

func (f *fileLoader) String() string {
	return fmt.Sprintf("loading from file %q", f.path)
}

type cliLoader struct{}

func (c *cliLoader) Load(ctx context.Context) (*googleoauth.Credentials, error) {
	return googleoauth.FindDefaultCredentials(ctx)
}

func (c *cliLoader) String() string {
	return fmt.Sprintf("loading from gcloud defaults")
}

type userLoader struct{}

func (u *userLoader) Load(ctx context.Context) (*googleoauth.Credentials, error) {
	var content string
	err := survey.Ask([]*survey.Question{
		{
			Prompt: &survey.Multiline{
				Message: "service account",
				Help:    "The location to file that contains the service account in JSON, or the service account in JSON format",
			},
		},
	}, &content)
	if err != nil {
		return nil, err
	}
	content = strings.TrimSpace(content)
	return (&fileOrContentLoader{content: content}).Load(ctx)
}
