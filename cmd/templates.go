package cmd

// Get-command template strings. Each const is rendered by renderGet() against
// the resource-specific view struct defined in the same file as its command.

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
