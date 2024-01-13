/*
Copyright 2017 The Kubernetes Authors.

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

package gcetasks

import (
	"context"
	"fmt"
	"reflect"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// ForwardingRule represents a GCE ForwardingRule
// +kops:fitask
type ForwardingRule struct {
	Name      *string
	Lifecycle fi.Lifecycle

	PortRange  *string
	Ports      []string
	TargetPool *TargetPool
	// An IP address can be specified either in dotted decimal
	// or by reference to an address object.  The following two
	// fields are mutually exclusive.
	IPAddress     *Address
	RuleIPAddress *string

	IPProtocol          string
	LoadBalancingScheme *string
	Network             *Network
	Subnetwork          *Subnet
	BackendService      *BackendService

	// Labels to set on the resource.
	Labels map[string]string

	// Fingerprint of the labels, used to avoid race-conditions on updates.
	// Only set on the actual resource returned by Find.
	labelFingerprint string

	// pruneForwardingRules will prune any forwarding rules found with the specified names
	pruneForwardingRules []forwardingRulePruneSpec
}

type forwardingRulePruneSpec struct {
	Name string
}

var _ fi.CompareWithID = &ForwardingRule{}

func (e *ForwardingRule) CompareWithID() *string {
	return e.Name
}

func (e *ForwardingRule) PruneForwardingRulesWithName(name string) {
	e.pruneForwardingRules = append(e.pruneForwardingRules, forwardingRulePruneSpec{Name: name})
}

func (e *ForwardingRule) Find(c *fi.CloudupContext) (*ForwardingRule, error) {
	ctx := c.Context()

	cloud := c.T.Cloud.(gce.GCECloud)
	name := fi.ValueOf(e.Name)

	r, err := cloud.Compute().ForwardingRules().Get(ctx, cloud.Project(), cloud.Region(), name)
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting ForwardingRule %q: %v", name, err)
	}

	actual := &ForwardingRule{
		Name:       fi.PtrTo(r.Name),
		IPProtocol: r.IPProtocol,
	}
	if r.PortRange != "" {
		actual.PortRange = &r.PortRange
	}
	if len(r.Ports) > 0 {
		actual.Ports = r.Ports
	}

	if r.Target != "" {
		actual.TargetPool = &TargetPool{
			Name: fi.PtrTo(lastComponent(r.Target)),
		}
	}
	if r.IPAddress != "" {
		address, err := findAddressByIP(cloud, r.IPAddress)
		if err != nil {
			return nil, fmt.Errorf("error finding Address with IP=%q: %v", r.IPAddress, err)
		}
		actual.IPAddress = address
	}
	if r.BackendService != "" {
		actual.BackendService = &BackendService{
			Name: fi.PtrTo(lastComponent(r.BackendService)),
		}
	}
	if r.LoadBalancingScheme != "" {
		actual.LoadBalancingScheme = fi.PtrTo(r.LoadBalancingScheme)
	}
	if r.Network != "" {
		actual.Network = &Network{
			Name: fi.PtrTo(lastComponent(r.Network)),
		}
	}
	if r.Subnetwork != "" {
		actual.Subnetwork = &Subnet{
			Name: fi.PtrTo(lastComponent(r.Subnetwork)),
		}
	}

	actual.Labels = r.Labels
	actual.labelFingerprint = r.LabelFingerprint

	// Ignore "system" fields
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *ForwardingRule) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
}

func (_ *ForwardingRule) CheckChanges(a, e, changes *ForwardingRule) error {
	if fi.ValueOf(e.Name) == "" {
		return fi.RequiredField("Name")
	}
	return nil
}

func (_ *ForwardingRule) RenderGCE(t *gce.GCEAPITarget, a, e, changes *ForwardingRule) error {
	ctx := context.TODO()

	name := fi.ValueOf(e.Name)

	o := &compute.ForwardingRule{
		Name:       name,
		IPProtocol: e.IPProtocol,
	}
	if e.PortRange != nil {
		o.PortRange = *e.PortRange
	}
	if len(e.Ports) > 0 {
		o.Ports = e.Ports
	}

	if e.LoadBalancingScheme != nil {
		o.LoadBalancingScheme = *e.LoadBalancingScheme
	}

	if e.TargetPool != nil {
		o.Target = e.TargetPool.URL(t.Cloud)
	}

	if e.BackendService != nil {
		if o.Target != "" {
			return fmt.Errorf("cannot specify both %q and %q for forwarding rule target.", o.Target, e.BackendService)
		}
		o.BackendService = e.BackendService.URL(t.Cloud)
	}

	if e.IPAddress != nil {
		o.IPAddress = fi.ValueOf(e.IPAddress.IPAddress)
		if o.IPAddress == "" {
			addr, err := e.IPAddress.find(t.Cloud)
			if err != nil {
				return fmt.Errorf("error finding Address %q: %v", e.IPAddress, err)
			}
			if addr == nil {
				return fmt.Errorf("Address %q was not found", e.IPAddress)
			}

			o.IPAddress = fi.ValueOf(addr.IPAddress)
			if o.IPAddress == "" {
				return fmt.Errorf("Address had no IP: %v", e.IPAddress)
			}
		}
	}
	if o.IPAddress != "" && e.RuleIPAddress != nil {
		return fmt.Errorf("Specified both IP Address and rule-managed IP address: %v, %v", e.IPAddress, *e.RuleIPAddress)
	}
	if e.RuleIPAddress != nil {
		o.IPAddress = *e.RuleIPAddress
	}

	if e.Network != nil {
		project := t.Cloud.Project()
		if e.Network.Project != nil {
			project = *e.Network.Project
		}
		o.Network = e.Network.URL(project)
	}

	if e.Subnetwork != nil {
		project := t.Cloud.Project()
		if e.Network.Project != nil {
			project = *e.Network.Project
		}
		o.Subnetwork = e.Subnetwork.URL(project, t.Cloud.Region())
	}

	if a == nil {
		klog.V(4).Infof("Creating ForwardingRule %q", o.Name)

		op, err := t.Cloud.Compute().ForwardingRules().Insert(ctx, t.Cloud.Project(), t.Cloud.Region(), o)
		if err != nil {
			return fmt.Errorf("error creating ForwardingRule %q: %v", o.Name, err)
		}

		if err := t.Cloud.WaitForOp(op); err != nil {
			return fmt.Errorf("error creating forwarding rule: %v", err)
		}

		if e.Labels != nil {
			// We can't set labels on creation; we have to read the object to get the fingerprint
			// TODO: We could get it from the operation!
			r, err := t.Cloud.Compute().ForwardingRules().Get(ctx, t.Cloud.Project(), t.Cloud.Region(), name)
			if err != nil {
				return fmt.Errorf("reading created ForwardingRule %q: %v", name, err)
			}

			req := compute.RegionSetLabelsRequest{
				LabelFingerprint: r.LabelFingerprint,
				Labels:           e.Labels,
			}
			op, err := t.Cloud.Compute().ForwardingRules().SetLabels(ctx, t.Cloud.Project(), t.Cloud.Region(), o.Name, &req)
			if err != nil {
				return fmt.Errorf("setting ForwardingRule labels: %w", err)
			}

			if err := t.Cloud.WaitForOp(op); err != nil {
				return fmt.Errorf("setting ForwardRule labels: %w", err)
			}
		}
	} else {
		if changes.Labels != nil {
			req := compute.RegionSetLabelsRequest{
				LabelFingerprint: a.labelFingerprint,
				Labels:           e.Labels,
			}
			op, err := t.Cloud.Compute().ForwardingRules().SetLabels(ctx, t.Cloud.Project(), t.Cloud.Region(), o.Name, &req)
			if err != nil {
				return fmt.Errorf("setting ForwardingRule labels: %w", err)
			}

			if err := t.Cloud.WaitForOp(op); err != nil {
				return fmt.Errorf("setting ForwardRule labels: %w", err)
			}

			changes.Labels = nil
		}

		if !reflect.DeepEqual(changes, &ForwardingRule{}) {
			return fmt.Errorf("cannot apply changes to ForwardingRule: %v", changes)
		}
	}

	return nil
}

type terraformForwardingRule struct {
	Name                string                   `cty:"name"`
	PortRange           *string                  `cty:"port_range"`
	Ports               []string                 `cty:"ports"`
	Target              *terraformWriter.Literal `cty:"target"`
	IPAddress           *terraformWriter.Literal `cty:"ip_address"`
	IPProtocol          string                   `cty:"ip_protocol"`
	LoadBalancingScheme *string                  `cty:"load_balancing_scheme"`
	Network             *terraformWriter.Literal `cty:"network"`
	Subnetwork          *terraformWriter.Literal `cty:"subnetwork"`
	BackendService      *terraformWriter.Literal `cty:"backend_service"`
	Labels              map[string]string        `cty:"labels"`
}

func (_ *ForwardingRule) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *ForwardingRule) error {
	name := fi.ValueOf(e.Name)

	tf := &terraformForwardingRule{
		Name:                name,
		IPProtocol:          e.IPProtocol,
		LoadBalancingScheme: e.LoadBalancingScheme,
		Ports:               e.Ports,
		PortRange:           e.PortRange,
		Labels:              e.Labels,
	}

	if e.TargetPool != nil {
		tf.Target = e.TargetPool.TerraformLink()
	}

	if e.Network != nil {
		tf.Network = e.Network.TerraformLink()
	}

	if e.Subnetwork != nil {
		tf.Subnetwork = e.Subnetwork.TerraformLink()
	}

	if e.BackendService != nil {
		tf.BackendService = e.BackendService.TerraformAddress()
	}

	if e.IPAddress != nil {
		tf.IPAddress = e.IPAddress.TerraformAddress()
	} else if e.RuleIPAddress != nil {
		tf.IPAddress = terraformWriter.LiteralFromStringValue(*e.RuleIPAddress)
	}

	return t.RenderResource("google_compute_forwarding_rule", name, tf)
}

func (e *ForwardingRule) TerraformLink() *terraformWriter.Literal {
	name := fi.ValueOf(e.Name)

	return terraformWriter.LiteralSelfLink("google_compute_forwarding_rule", name)
}

var _ fi.CloudupProducesDeletions = &ForwardingRule{}

// FindDeletions implements fi.HasDeletions
func (e *ForwardingRule) FindDeletions(c *fi.CloudupContext) ([]fi.CloudupDeletion, error) {
	var removals []fi.CloudupDeletion

	if len(e.pruneForwardingRules) != 0 {
		ctx := c.Context()
		cloud := c.T.Cloud.(gce.GCECloud)

		forwardingRules, err := cloud.Compute().ForwardingRules().List(ctx, cloud.Project(), cloud.Region())
		if err != nil {
			return nil, fmt.Errorf("listing forwardingRules: %w", err)
		}

		for _, forwardingRule := range forwardingRules {
			prune := false

			for _, rule := range e.pruneForwardingRules {
				if rule.Name == forwardingRule.Name {
					prune = true
				}
			}

			if prune {
				removals = append(removals, &deleteForwardingRule{forwardingRule: forwardingRule})
			}
		}
	}

	return removals, nil
}

// deleteForwardingRule tracks a ForwardingRule that we're going to delete
// It implements fi.Deletion
type deleteForwardingRule struct {
	forwardingRule *compute.ForwardingRule
}

var _ fi.CloudupDeletion = &deleteForwardingRule{}

// TaskName returns the task name
func (d *deleteForwardingRule) TaskName() string {
	return "ForwardingRule"
}

// Item returns the forwarding rule name
func (d *deleteForwardingRule) Item() string {
	return d.forwardingRule.Name
}

func (d *deleteForwardingRule) Delete(t fi.CloudupTarget) error {
	ctx := context.TODO()

	gceTarget, ok := t.(*gce.GCEAPITarget)
	if !ok {
		return fmt.Errorf("unexpected target type for deletion: %T", t)
	}

	cloud := gceTarget.Cloud
	name := d.forwardingRule.Name

	if _, err := cloud.Compute().ForwardingRules().Delete(ctx, cloud.Project(), cloud.Region(), name); err != nil {
		return fmt.Errorf("deleting forwardingRule %s: %w", name, err)
	}

	return nil
}

// String returns a string representation of the task
func (d *deleteForwardingRule) String() string {
	return d.TaskName() + "-" + d.Item()
}

func (d *deleteForwardingRule) DeferDeletion() bool {
	// We want to defer deletion, in case new nodes are launched with
	// the old configuration during the rolling-update operation.
	return true
}
