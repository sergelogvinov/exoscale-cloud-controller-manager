package exoscale

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"sync"
	"time"
)

var (
	testAPIKey    = new(exoscaleCCMTestSuite).randomString(10)
	testAPISecret = new(exoscaleCCMTestSuite).randomString(10)
)

func (ts *exoscaleCCMTestSuite) Test_newRefreshableExoscaleClient_no_env() {
	os.Unsetenv("EXOSCALE_API_KEY")
	os.Unsetenv("EXOSCALE_API_SECRET")

	_, err := newRefreshableExoscaleClient(context.Background())
	ts.Require().Error(err)
}

func (ts *exoscaleCCMTestSuite) Test_newRefreshableExoscaleClient_env_credentials() {
	os.Setenv("EXOSCALE_API_KEY", testAPIKey)
	os.Setenv("EXOSCALE_API_SECRET", testAPISecret)
	defer func() {
		os.Unsetenv("EXOSCALE_API_KEY")
		os.Unsetenv("EXOSCALE_API_SECRET")
	}()

	expected := &refreshableExoscaleClient{
		RWMutex: &sync.RWMutex{},
		apiCredentials: exoscaleAPICredentials{
			APIKey:    testAPIKey,
			APISecret: testAPISecret,
		},
		apiEnvironment: defaultComputeEnvironment,
	}

	actual, err := newRefreshableExoscaleClient(context.Background())
	ts.Require().NoError(err)
	ts.Require().Equal(expected.apiCredentials, actual.apiCredentials)
	ts.Require().NotNil(actual.exo)
}

func (ts *exoscaleCCMTestSuite) Test_refreshableExoscaleClient_refreshCredentials() {
	testAPICredentials := exoscaleAPICredentials{
		APIKey:    testAPISecret,
		APISecret: testAPISecret,
		Name:      ts.randomString(10),
	}

	jsonAPICredentials, err := json.Marshal(testAPICredentials)
	ts.Require().NoError(err)

	tmpdir, err := os.MkdirTemp(os.TempDir(), "exoscale-ccm")
	ts.Require().NoError(err)
	defer os.RemoveAll(tmpdir)

	testAPICredentialsFile := path.Join(tmpdir, "credentials.json")

	ts.Require().NoError(os.WriteFile(testAPICredentialsFile, jsonAPICredentials, 0o600))

	client := &refreshableExoscaleClient{RWMutex: &sync.RWMutex{}}
	client.refreshCredentialsFromFile(testAPICredentialsFile)

	client.RLock()
	defer client.RUnlock()
	ts.Require().Equal(testAPICredentials, client.apiCredentials)
	ts.Require().NotNil(client.exo)
}

func (ts *exoscaleCCMTestSuite) Test_refreshableExoscaleClient_watchCredentialsFile() {
	testAPICredentials := exoscaleAPICredentials{
		APIKey:    testAPISecret,
		APISecret: testAPISecret,
		Name:      ts.randomString(10),
	}

	jsonAPICredentials, err := json.Marshal(testAPICredentials)
	ts.Require().NoError(err)

	tmpdir, err := os.MkdirTemp(os.TempDir(), "exoscale-ccm")
	ts.Require().NoError(err)
	defer os.RemoveAll(tmpdir)

	testAPICredentialsFile := path.Join(tmpdir, "credentials.json")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := &refreshableExoscaleClient{RWMutex: &sync.RWMutex{}}
	go client.watchCredentialsFile(ctx, testAPICredentialsFile)

	time.Sleep(1 * time.Second)
	ts.Require().NoError(os.WriteFile(testAPICredentialsFile, jsonAPICredentials, 0o600))
	time.Sleep(1 * time.Second)

	client.RLock()
	defer client.RUnlock()
	ts.Require().Equal(client.apiCredentials, testAPICredentials)
	ts.Require().NotNil(client.exo)
}
