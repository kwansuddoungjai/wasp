/*
Wasp API

REST API for the Wasp node

API version: 123
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package apiclient

import (
	"encoding/json"
)

// checks if the NodeConnectionMetrics type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &NodeConnectionMetrics{}

// NodeConnectionMetrics struct for NodeConnectionMetrics
type NodeConnectionMetrics struct {
	InMilestone *NodeConnectionMessageMetrics `json:"inMilestone,omitempty"`
	NodeConnectionMessagesMetrics *NodeConnectionMessagesMetrics `json:"nodeConnectionMessagesMetrics,omitempty"`
	// Chain IDs of the chains registered to receiving L1 events
	Registered []string `json:"registered,omitempty"`
}

// NewNodeConnectionMetrics instantiates a new NodeConnectionMetrics object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewNodeConnectionMetrics() *NodeConnectionMetrics {
	this := NodeConnectionMetrics{}
	return &this
}

// NewNodeConnectionMetricsWithDefaults instantiates a new NodeConnectionMetrics object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewNodeConnectionMetricsWithDefaults() *NodeConnectionMetrics {
	this := NodeConnectionMetrics{}
	return &this
}

// GetInMilestone returns the InMilestone field value if set, zero value otherwise.
func (o *NodeConnectionMetrics) GetInMilestone() NodeConnectionMessageMetrics {
	if o == nil || isNil(o.InMilestone) {
		var ret NodeConnectionMessageMetrics
		return ret
	}
	return *o.InMilestone
}

// GetInMilestoneOk returns a tuple with the InMilestone field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *NodeConnectionMetrics) GetInMilestoneOk() (*NodeConnectionMessageMetrics, bool) {
	if o == nil || isNil(o.InMilestone) {
		return nil, false
	}
	return o.InMilestone, true
}

// HasInMilestone returns a boolean if a field has been set.
func (o *NodeConnectionMetrics) HasInMilestone() bool {
	if o != nil && !isNil(o.InMilestone) {
		return true
	}

	return false
}

// SetInMilestone gets a reference to the given NodeConnectionMessageMetrics and assigns it to the InMilestone field.
func (o *NodeConnectionMetrics) SetInMilestone(v NodeConnectionMessageMetrics) {
	o.InMilestone = &v
}

// GetNodeConnectionMessagesMetrics returns the NodeConnectionMessagesMetrics field value if set, zero value otherwise.
func (o *NodeConnectionMetrics) GetNodeConnectionMessagesMetrics() NodeConnectionMessagesMetrics {
	if o == nil || isNil(o.NodeConnectionMessagesMetrics) {
		var ret NodeConnectionMessagesMetrics
		return ret
	}
	return *o.NodeConnectionMessagesMetrics
}

// GetNodeConnectionMessagesMetricsOk returns a tuple with the NodeConnectionMessagesMetrics field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *NodeConnectionMetrics) GetNodeConnectionMessagesMetricsOk() (*NodeConnectionMessagesMetrics, bool) {
	if o == nil || isNil(o.NodeConnectionMessagesMetrics) {
		return nil, false
	}
	return o.NodeConnectionMessagesMetrics, true
}

// HasNodeConnectionMessagesMetrics returns a boolean if a field has been set.
func (o *NodeConnectionMetrics) HasNodeConnectionMessagesMetrics() bool {
	if o != nil && !isNil(o.NodeConnectionMessagesMetrics) {
		return true
	}

	return false
}

// SetNodeConnectionMessagesMetrics gets a reference to the given NodeConnectionMessagesMetrics and assigns it to the NodeConnectionMessagesMetrics field.
func (o *NodeConnectionMetrics) SetNodeConnectionMessagesMetrics(v NodeConnectionMessagesMetrics) {
	o.NodeConnectionMessagesMetrics = &v
}

// GetRegistered returns the Registered field value if set, zero value otherwise.
func (o *NodeConnectionMetrics) GetRegistered() []string {
	if o == nil || isNil(o.Registered) {
		var ret []string
		return ret
	}
	return o.Registered
}

// GetRegisteredOk returns a tuple with the Registered field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *NodeConnectionMetrics) GetRegisteredOk() ([]string, bool) {
	if o == nil || isNil(o.Registered) {
		return nil, false
	}
	return o.Registered, true
}

// HasRegistered returns a boolean if a field has been set.
func (o *NodeConnectionMetrics) HasRegistered() bool {
	if o != nil && !isNil(o.Registered) {
		return true
	}

	return false
}

// SetRegistered gets a reference to the given []string and assigns it to the Registered field.
func (o *NodeConnectionMetrics) SetRegistered(v []string) {
	o.Registered = v
}

func (o NodeConnectionMetrics) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o NodeConnectionMetrics) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.InMilestone) {
		toSerialize["inMilestone"] = o.InMilestone
	}
	if !isNil(o.NodeConnectionMessagesMetrics) {
		toSerialize["nodeConnectionMessagesMetrics"] = o.NodeConnectionMessagesMetrics
	}
	if !isNil(o.Registered) {
		toSerialize["registered"] = o.Registered
	}
	return toSerialize, nil
}

type NullableNodeConnectionMetrics struct {
	value *NodeConnectionMetrics
	isSet bool
}

func (v NullableNodeConnectionMetrics) Get() *NodeConnectionMetrics {
	return v.value
}

func (v *NullableNodeConnectionMetrics) Set(val *NodeConnectionMetrics) {
	v.value = val
	v.isSet = true
}

func (v NullableNodeConnectionMetrics) IsSet() bool {
	return v.isSet
}

func (v *NullableNodeConnectionMetrics) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableNodeConnectionMetrics(val *NodeConnectionMetrics) *NullableNodeConnectionMetrics {
	return &NullableNodeConnectionMetrics{value: val, isSet: true}
}

func (v NullableNodeConnectionMetrics) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableNodeConnectionMetrics) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


