// Copyright 2014-2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package audit

import (
	"fmt"
	"net/http"
	"time"

	log "github.com/cihub/seelog"
)

const (
	getCredentialsEventType       = "GetCredentials"
	getCredentialsAuditLogVersion = 1
)

type commonAuditLogEntryFields struct {
	eventTime    string
	responseCode int
	srcAddr      string
	theURL       string
	userAgent    string
}

func GetCredentialsEventType() string {
	return getCredentialsEventType
}

func (c *commonAuditLogEntryFields) string() string {
	return fmt.Sprintf("%s %d %s %s %s", c.eventTime, c.responseCode, c.srcAddr, c.theURL, c.userAgent)
}

type getCredentialsAuditLogEntryFields struct {
	eventType            string
	version              int
	cluster              string
	containerInstanceArn string
}

func (g *getCredentialsAuditLogEntryFields) string() string {
	return fmt.Sprintf("%s %d %s %s", g.eventType, g.version, g.cluster, g.containerInstanceArn)
}

func constructCommonAuditLogEntryFields(r *http.Request, httpResponseCode int) string {
	fields := &commonAuditLogEntryFields{
		eventTime:    time.Now().UTC().Format(time.RFC3339),
		responseCode: httpResponseCode,
		srcAddr:      populateField(r.RemoteAddr),
		theURL:       populateField(fmt.Sprintf(`"%s"`, r.URL.Path)),
		userAgent:    populateField(fmt.Sprintf(`"%s"`, r.UserAgent())),
	}
	return fields.string()
}

func constructAuditLogEntryByType(eventType string, cluster string, containerInstanceArn string) string {
	switch eventType {
	case getCredentialsEventType:
		fields := &getCredentialsAuditLogEntryFields{
			eventType:            eventType,
			version:              getCredentialsAuditLogVersion,
			cluster:              populateField(cluster),
			containerInstanceArn: populateField(containerInstanceArn),
		}
		return fields.string()
	default:
		log.Warn(fmt.Sprintf("Unknown eventType: %s", eventType))
		return ""
	}
}

func populateField(logField string) string {
	if logField == "" {
		logField = "-"
	}
	return logField
}