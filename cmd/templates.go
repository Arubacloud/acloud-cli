package cmd

// Get-command template strings. Each const is rendered by renderGet() against
// the resource-specific view struct defined in the same file as its command.
// Complex dynamic sections (DHCP routes, VPN IP config) are appended with
// fmt.Printf calls in the operation function after the template render.

const kmsGetTmpl = `
KMS Details:
============
ID:              {{.ID}}
URI:             {{.URI}}
Name:            {{.Name}}
Region:          {{.Region}}
Status:          {{.Status}}
Creation Date:   {{.CreatedAt}}
Created By:      {{.CreatedBy}}
Tags:            {{.Tags}}
`

const cloudServerGetTmpl = `
Cloud Server Details:
====================
ID:              {{.ID}}
Name:            {{.Name}}
Region:          {{.Region}}
Flavor:          {{.Flavor}}
CPU:             {{.CPU}}
RAM:             {{.RAM}}
HD:              {{.HD}}
Boot Volume URI: {{.BootVolumeURI}}
Keypair URI:     {{.KeypairURI}}
Status:          {{.Status}}
Tags:            {{.Tags}}
`

const keypairGetTmpl = `
Keypair Details:
===============
Name:            {{.Name}}
URI:             {{.URI}}
Public Key:      {{.PublicKey}}
Status:          Active
Creation Date:   {{.CreatedAt}}
Created By:      {{.CreatedBy}}
`

const projectGetTmpl = `
Project Details:
================
ID:              {{.ID}}
Name:            {{.Name}}
Description:     {{.Description}}
Default:         {{.Default}}
Creation Date:   {{.CreatedAt}}
Update Date:     {{.UpdatedAt}}
Created By:      {{.CreatedBy}}
Updated By:      {{.UpdatedBy}}
Tags:            {{.Tags}}
`

const containerRegistryGetTmpl = `
Container Registry Details:
==========================
ID:              {{.ID}}
URI:             {{.URI}}
Name:            {{.Name}}
Region:          {{.Region}}
Public IP:       {{.PublicIP}}
VPC:             {{.VPC}}
Subnet:          {{.Subnet}}
Security Group:  {{.SecurityGroup}}
Block Storage:   {{.BlockStorage}}
Billing Period:  {{.BillingPeriod}}
Admin User:      {{.AdminUser}}
Concurrent Users: {{.ConcurrentUsers}}
Status:          {{.Status}}
Creation Date:   {{.CreatedAt}}
Created By:      {{.CreatedBy}}
Tags:            {{.Tags}}
`

const elasticIPGetTmpl = `
Elastic IP Details:
===================
ID:              {{.ID}}
URI:             {{.URI}}
Name:            {{.Name}}
Region:          {{.Region}}
Address:         {{.Address}}
Billing Period:  {{.BillingPeriod}}
Linked Resources: {{.LinkedResources}}
Creation Date:   {{.CreatedAt}}
Created By:      {{.CreatedBy}}
Tags:            {{.Tags}}
Status:          {{.Status}}
`

const loadBalancerGetTmpl = `
Load Balancer Details:
======================
ID:              {{.ID}}
URI:             {{.URI}}
Name:            {{.Name}}
Address:         {{.Address}}
VPC:             {{.VPC}}
Linked Resources: {{.LinkedResources}}
Creation Date:   {{.CreatedAt}}
Created By:      {{.CreatedBy}}
Tags:            {{.Tags}}
Status:          {{.Status}}
`

const securityGroupGetTmpl = `
Security Group Details:
======================
ID:              {{.ID}}
URI:             {{.URI}}
Name:            {{.Name}}
Region:          {{.Region}}
Creation Date:   {{.CreatedAt}}
Created By:      {{.CreatedBy}}
Tags:            {{.Tags}}
Status:          {{.Status}}
`

const securityRuleGetTmpl = `
Security Rule Details:
=====================
ID:              {{.ID}}
URI:             {{.URI}}
Name:            {{.Name}}
Region:          {{.Region}}
Direction:       {{.Direction}}
Protocol:        {{.Protocol}}
Port:            {{.Port}}
Target Kind:     {{.TargetKind}}
Target Value:    {{.TargetValue}}
Creation Date:   {{.CreatedAt}}
Created By:      {{.CreatedBy}}
Tags:            {{.Tags}}
Status:          {{.Status}}
`

const subnetGetTmpl = `
Subnet Details:
===============
ID:              {{.ID}}
URI:             {{.URI}}
Name:            {{.Name}}
Region:          {{.Region}}
Type:            {{.Type}}
CIDR:            {{.CIDR}}
`

const vpcGetTmpl = `
VPC Details:
============
ID:              {{.ID}}
URI:             {{.URI}}
Name:            {{.Name}}
Region:          {{.Region}}
Default:         {{.Default}}
Linked Resources: {{.LinkedResources}}
Creation Date:   {{.CreatedAt}}
Created By:      {{.CreatedBy}}
Tags:            {{.Tags}}
Status:          {{.Status}}
`

const vpcPeeringGetTmpl = `
VPC Peering Details:
====================
ID:              {{.ID}}
Name:            {{.Name}}
Peer VPC:        {{.PeerVPC}}
Region:          {{.Region}}
Creation Date:   {{.CreatedAt}}
Created By:      {{.CreatedBy}}
Tags:            {{.Tags}}
Status:          {{.Status}}
`

const vpcPeeringRouteGetTmpl = `
VPC Peering Route Details:
==========================
ID:              {{.ID}}
Name:            {{.Name}}
Local Network:    {{.LocalNetwork}}
Remote Network:   {{.RemoteNetwork}}
Billing Period:  {{.BillingPeriod}}
Tags:            {{.Tags}}
Status:          {{.Status}}
`

const vpnRouteGetTmpl = `
VPN Route Details:
==================
ID:              {{.ID}}
URI:             {{.URI}}
Name:            {{.Name}}
Region:          {{.Region}}
Cloud Subnet:    {{.CloudSubnet}}
OnPrem Subnet:   {{.OnPremSubnet}}
Creation Date:   {{.CreatedAt}}
Created By:      {{.CreatedBy}}
Tags:            {{.Tags}}
Status:          {{.Status}}
`

const vpnTunnelGetTmpl = `
VPN Tunnel Details:
===================
ID:              {{.ID}}
URI:             {{.URI}}
Name:            {{.Name}}
Region:          {{.Region}}
VPN Type:        {{.VPNType}}
Protocol:        {{.Protocol}}
Peer IP:         {{.PeerIP}}
`

const scheduleJobGetTmpl = `
Job Details:
============
ID:              {{.ID}}
URI:             {{.URI}}
Name:            {{.Name}}
Region:          {{.Region}}
Job Type:        {{.JobType}}
Enabled:         {{.Enabled}}
Schedule At:     {{.ScheduleAt}}
CRON:            {{.Cron}}
Execute Until:   {{.ExecuteUntil}}
Status:          {{.Status}}
Creation Date:   {{.CreatedAt}}
Created By:      {{.CreatedBy}}
Tags:            {{.Tags}}
`

const storageBackupGetTmpl = `
Storage Backup Details:
=======================
ID:              {{.ID}}
URI:             {{.URI}}
Name:            {{.Name}}
Type:            {{.Type}}
Source Volume:   {{.SourceVolume}}
Retention Days:  {{.RetentionDays}}
Billing Period:  {{.BillingPeriod}}
Region:          {{.Region}}
Status:          {{.Status}}
Creation Date:   {{.CreatedAt}}
Created By:      {{.CreatedBy}}
Tags:            {{.Tags}}
`

const snapshotGetTmpl = `
Snapshot Details:
=================
ID:              {{.ID}}
URI:             {{.URI}}
Name:            {{.Name}}
Size (GB):       {{.Size}}
Source Volume:   {{.SourceVolume}}
Region:          {{.Region}}
Status:          {{.Status}}
Creation Date:   {{.CreatedAt}}
Created By:      {{.CreatedBy}}
Tags:            {{.Tags}}
`

const restoreGetTmpl = `
Restore Operation Details:
==========================
ID:              {{.ID}}
URI:             {{.URI}}
Name:            {{.Name}}
Target Volume:   {{.TargetVolume}}
Region:          {{.Region}}
Status:          {{.Status}}
Creation Date:   {{.CreatedAt}}
Created By:      {{.CreatedBy}}
Tags:            {{.Tags}}
`

const dbaasGetTmpl = `
DBaaS Instance Details:
=======================
ID:              {{.ID}}
URI:             {{.URI}}
Name:            {{.Name}}
Region:          {{.Region}}
Engine Type:     {{.EngineType}}
Engine Version:  {{.EngineVersion}}
Engine Name:     {{.EngineName}}
Flavor:          {{.Flavor}}
Status:          {{.Status}}
Creation Date:   {{.CreatedAt}}
Created By:      {{.CreatedBy}}
Tags:            {{.Tags}}
`

const databaseGetTmpl = `
Database Details:
================
Name:            {{.Name}}
Creation Date:   {{.CreatedAt}}
Created By:      {{.CreatedBy}}
`

const userGetTmpl = `
User Details:
=============
Username:        {{.Username}}
Creation Date:   {{.CreatedAt}}
Created By:      {{.CreatedBy}}
`

const databaseBackupGetTmpl = `
Backup Details:
==============
ID:              {{.ID}}
URI:             {{.URI}}
Name:            {{.Name}}
Region:          {{.Region}}
Status:          {{.Status}}
Creation Date:   {{.CreatedAt}}
Created By:      {{.CreatedBy}}
Tags:            {{.Tags}}
`

const kaasGetTmpl = `
KaaS Cluster Details:
====================
ID:                 {{.ID}}
URI:                {{.URI}}
Name:               {{.Name}}
Region:             {{.Region}}
Kubernetes Version: {{.Version}}
Status:             {{.Status}}
Creation Date:      {{.CreatedAt}}
Created By:         {{.CreatedBy}}
Tags:               {{.Tags}}
`

const blockStorageGetTmpl = `
Block Storage Details:
======================
ID:              {{.ID}}
URI:             {{.URI}}
Name:            {{.Name}}
Size (GB):       {{.Size}}
Type:            {{.Type}}
Zone:            {{.Zone}}
Region:          {{.Region}}
Bootable:        {{.Bootable}}
Status:          {{.Status}}
Creation Date:   {{.CreatedAt}}
Created By:      {{.CreatedBy}}
Tags:            {{.Tags}}
`
