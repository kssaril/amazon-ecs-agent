// Copyright 2014-2015 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package dockerclient

import (
	"fmt"
	"testing"

	"github.com/aws/amazon-ecs-agent/agent/engine/dockerclient/mocks"
	"github.com/golang/mock/gomock"
)

func TestGetDefaultClientSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_dockerclient.NewMockClient(ctrl)
	mockClient.EXPECT().Ping()

	expectedEndpoint := "expectedEndpoint"

	newVersionedClient = func(endpoint, version string) (Client, error) {
		if endpoint != expectedEndpoint {
			t.Errorf("Expected endpoint %s but was %s", expectedEndpoint, endpoint)
		}
		if version != string(defaultVersion) {
			t.Errorf("Expected version %s but was %s", defaultVersion, version)
		}
		return mockClient, nil
	}

	factory := NewFactory(expectedEndpoint)
	client, err := factory.GetDefaultClient()
	if err != nil {
		t.Fatal("err should be nil")
	}
	if client != mockClient {
		t.Error("Client returned by GetDefaultClient differs from mockClient")
	}
}

func TestGetClientCached(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_dockerclient.NewMockClient(ctrl)
	mockClient.EXPECT().Ping()

	expectedEndpoint := "expectedEndpoint"

	newVersionedClient = func(endpoint, version string) (Client, error) {
		if endpoint != expectedEndpoint {
			t.Errorf("Expected endpoint %s but was %s", expectedEndpoint, endpoint)
		}
		if version != string(version_1_18) {
			t.Errorf("Expected version %s but was %s", version_1_18, version)
		}
		return mockClient, nil
	}

	factory := NewFactory(expectedEndpoint)
	client, err := factory.GetClient(version_1_18)
	if err != nil {
		t.Fatal("err should be nil")
	}
	if client != mockClient {
		t.Error("Client returned by GetDefaultClient differs from mockClient")
	}

	client, err = factory.GetClient(version_1_18)
	if err != nil {
		t.Fatal("err should be nil")
	}
	if client != mockClient {
		t.Error("Client returned by GetDefaultClient differs from mockClient")
	}
}

func TestGetClientFailCreateNotCached(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_dockerclient.NewMockClient(ctrl)

	calledOnce := false
	newVersionedClient = func(endpoint, version string) (Client, error) {
		if calledOnce {
			return mockClient, nil
		}
		calledOnce = true
		return mockClient, fmt.Errorf("Test error!")
	}

	factory := NewFactory("")
	client, err := factory.GetClient(version_1_19)
	if err == nil {
		t.Fatal("err should not be nil")
	}
	if client != nil {
		t.Error("client should be nil")
	}

	mockClient.EXPECT().Ping()

	client, err = factory.GetClient(version_1_19)
	if err != nil {
		t.Fatal("err should be nil")
	}
	if client != mockClient {
		t.Error("Client returned by GetDefaultClient differs from mockClient")
	}
}

func TestGetClientFailPingNotCached(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock_dockerclient.NewMockClient(ctrl)

	newVersionedClient = func(endpoint, version string) (Client, error) {
		return mockClient, nil
	}

	mockClient.EXPECT().Ping().Return(fmt.Errorf("Test error!"))

	factory := NewFactory("")
	client, err := factory.GetClient(version_1_20)
	if err == nil {
		t.Fatal("err should not be nil")
	}
	if client != nil {
		t.Error("client should be nil")
	}

	mockClient.EXPECT().Ping()

	client, err = factory.GetClient(version_1_20)
	if err != nil {
		t.Fatal("err should be nil")
	}
	if client != mockClient {
		t.Error("Client returned by GetDefaultClient differs from mockClient")
	}
}

func TestFindAvailableVersiosn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient117 := mock_dockerclient.NewMockClient(ctrl)
	mockClient118 := mock_dockerclient.NewMockClient(ctrl)
	mockClient119 := mock_dockerclient.NewMockClient(ctrl)
	mockClient120 := mock_dockerclient.NewMockClient(ctrl)

	expectedEndpoint := "expectedEndpoint"

	newVersionedClient = func(endpoint, version string) (Client, error) {
		if endpoint != expectedEndpoint {
			t.Errorf("Expected endpoint %s but was %s", expectedEndpoint, endpoint)
		}

		switch dockerVersion(version) {
		case version_1_17:
			return mockClient117, nil
		case version_1_18:
			return mockClient118, nil
		case version_1_19:
			return mockClient119, nil
		case version_1_20:
			return mockClient120, nil
		default:
			t.Fatal("Unrecognized version")
		}
		return nil, fmt.Errorf("This should not happen, update the test")
	}

	mockClient117.EXPECT().Ping()
	mockClient118.EXPECT().Ping().Return(fmt.Errorf("Test error!"))
	mockClient119.EXPECT().Ping()
	mockClient120.EXPECT().Ping()

	expectedVersions := []dockerVersion{version_1_17, version_1_19, version_1_20}

	factory := NewFactory(expectedEndpoint)
	versions := factory.FindAvailableVersions()
	if len(versions) != len(expectedVersions) {
		t.Errorf("Expected %d versions but got %d", len(expectedVersions), len(versions))
	}
	for i := 0; i < len(versions); i++ {
		if versions[i] != expectedVersions[i] {
			t.Errorf("Expected version %s but got version %s", expectedVersions[i], versions[i])
		}
	}
}