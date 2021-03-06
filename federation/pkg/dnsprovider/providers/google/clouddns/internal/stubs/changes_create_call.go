/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package stubs

import (
	"fmt"

	"google.golang.org/api/googleapi"
	"k8s.io/kubernetes/federation/pkg/dnsprovider/providers/google/clouddns/internal/interfaces"
)

// Compile time check for interface adeherence
var _ interfaces.ChangesCreateCall = ChangesCreateCall{}

type ChangesCreateCall struct {
	Service *ChangesService
	Project string
	Zone    string
	Change  interfaces.Change
	Error   error // Use this to over-ride response if necessary
}

func hashKey(set interfaces.ResourceRecordSet) string {
	return fmt.Sprintf("%s-%d-%s", set.Name(), set.Ttl(), string(set.Type()))
}

func (c ChangesCreateCall) Do(opts ...googleapi.CallOption) (interfaces.Change, error) {
	if c.Error != nil {
		return nil, c.Error
	}
	zone := (c.Service.Service.ManagedZones_.Impl[c.Project][c.Zone]).(*ManagedZone)
	rrsets := map[string]int{} // Simple mechanism to detect dupes and missing rrsets before committing - stores index+1
	for i, set := range zone.Rrsets {
		rrsets[hashKey(set)] = i + 1
	}
	for _, add := range c.Change.Additions() {
		if rrsets[hashKey(add)] > 0 {
			return nil, fmt.Errorf("Attempt to insert duplicate rrset %v", add)
		}
	}
	for _, del := range c.Change.Deletions() {
		if !(rrsets[hashKey(del)] > 0) {
			return nil, fmt.Errorf("Attempt to delete non-existent rrset %v", del)
		}
	}
	for _, add := range c.Change.Additions() {
		zone.Rrsets = append(zone.Rrsets, *(add.(*ResourceRecordSet)))
	}
	for _, del := range c.Change.Deletions() {
		zone.Rrsets = append(
			zone.Rrsets[:rrsets[hashKey(del)]-1],
			zone.Rrsets[rrsets[hashKey(del)]:]...)
	}
	return c.Change, nil
}
