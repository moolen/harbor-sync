// +build !ignore_autogenerated

/*
Copyright 2019 The Authors.

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

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HarborSync) DeepCopyInto(out *HarborSync) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HarborSync.
func (in *HarborSync) DeepCopy() *HarborSync {
	if in == nil {
		return nil
	}
	out := new(HarborSync)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HarborSync) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HarborSyncList) DeepCopyInto(out *HarborSyncList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]HarborSync, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HarborSyncList.
func (in *HarborSyncList) DeepCopy() *HarborSyncList {
	if in == nil {
		return nil
	}
	out := new(HarborSyncList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HarborSyncList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HarborSyncSpec) DeepCopyInto(out *HarborSyncSpec) {
	*out = *in
	if in.Mapping != nil {
		in, out := &in.Mapping, &out.Mapping
		*out = make([]ProjectMapping, len(*in))
		copy(*out, *in)
	}
	if in.Webhook != nil {
		in, out := &in.Webhook, &out.Webhook
		*out = make([]WebhookConfig, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HarborSyncSpec.
func (in *HarborSyncSpec) DeepCopy() *HarborSyncSpec {
	if in == nil {
		return nil
	}
	out := new(HarborSyncSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HarborSyncStatus) DeepCopyInto(out *HarborSyncStatus) {
	*out = *in
	if in.RobotCredentials != nil {
		in, out := &in.RobotCredentials, &out.RobotCredentials
		*out = make(map[string]RobotAccountCredentials, len(*in))
		for key, val := range *in {
			var outVal []RobotAccountCredential
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = make(RobotAccountCredentials, len(*in))
				copy(*out, *in)
			}
			(*out)[key] = outVal
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HarborSyncStatus.
func (in *HarborSyncStatus) DeepCopy() *HarborSyncStatus {
	if in == nil {
		return nil
	}
	out := new(HarborSyncStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProjectMapping) DeepCopyInto(out *ProjectMapping) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProjectMapping.
func (in *ProjectMapping) DeepCopy() *ProjectMapping {
	if in == nil {
		return nil
	}
	out := new(ProjectMapping)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RobotAccountCredential) DeepCopyInto(out *RobotAccountCredential) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RobotAccountCredential.
func (in *RobotAccountCredential) DeepCopy() *RobotAccountCredential {
	if in == nil {
		return nil
	}
	out := new(RobotAccountCredential)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in RobotAccountCredentials) DeepCopyInto(out *RobotAccountCredentials) {
	{
		in := &in
		*out = make(RobotAccountCredentials, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RobotAccountCredentials.
func (in RobotAccountCredentials) DeepCopy() RobotAccountCredentials {
	if in == nil {
		return nil
	}
	out := new(RobotAccountCredentials)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebhookConfig) DeepCopyInto(out *WebhookConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebhookConfig.
func (in *WebhookConfig) DeepCopy() *WebhookConfig {
	if in == nil {
		return nil
	}
	out := new(WebhookConfig)
	in.DeepCopyInto(out)
	return out
}
