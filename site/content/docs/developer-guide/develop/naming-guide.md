---
title: Plugin Naming Guide
slug: naming-guide
weight: 200
---

This document describes guidelines for naming your kubectl plugins.

These guidelines are used for reviewing the plugins [submitted to Krew]({{<ref
"../release/submitting-to-krew.md">}}).

##### _Use lowercase and hyphens_

Plugin names must be all lowercase and separate words with hyphens.
**Don't** use camelCase, PascalCase, or snake_case; use
[kebab-case](http://wiki.c2.com/?KebabCase).

- **NO:** `kubectl OpenSvc`
- **YES:** `kubectl open-svc`

##### _Be specific_

Plugin names should not be verbs or nouns that are generic, already overloaded, or
likely to be used for broader purposes by another plugin.

- **NO:** `kubectl login` (Too broad)
- **YES:** `kubectl gke-login`

Also:

- **NO:** `kubectl ui` (Should be used only for Kubernetes Dashboard)
- **YES:** `kubectl gke-ui`

##### _Be unique_

Find a unique name for your plugin that differentiates it from other
plugins that perform a similar function.

- **NO:** `kubectl view-logs` (Unclear how it is different from the builtin
  "logs" command, or many other tools for viewing logs)
- **YES:** `kubectl tailer` (Unique name, points to the underlying)
  tool name.

##### _Use Verbs and Resource Types_

If the name does not make it clear (a) what verb the plugin is doing on a
resource, or (b) what kind of resource it's doing the action on, consider
clarifying unless it is obvious.

- **NO:** `kubectl service` (Unclear what this plugin is doing with)
  service.
- **NO:** `kubectl open` (Unclear what the plugin is opening)
- **YES:** `kubectl open-svc` (It is clear the plugin will open a service)

##### _Prefix Vendor Identifiers_

Use vendor-specific strings as prefix, separated with a dash. This makes it
easier to search/group plugins that are about a specific vendor.

- **NO:** `kubectl ui-gke` (Makes it harder to search or locate in a
  plugin list)
- **YES:** `kubectl gke-ui` (Will show up together with other gke-* plugins)

##### _Avoid repeating kube[rnetes]_

Plugin names should not include "kube-" or "kubernetes-" prefixes.

- **NO:** `kubectl kube-node-admin` ("kubectl " already has "kube" in it)
- **YES:** `kubectl node-admin`
  
While it is not recommended to include "kube*" in the plugin command name it
is recommended that the code repository starts with "kubectl-" so plugin
source code can be found outside of krew and the intended use is clear.

##### _Avoid Resource Acronyms and Abbreviations_

Using kubectl acronyms for API resources (e.g. svc, ing, deploy, cm) reduces
the readability and discoverability of a plugin, which is more important
than the few keystrokes saved.

- **NO:** `kubectl new-ing` (Unclear that the plugin is for Ingress)
- **YES:** `kubectl debug-ingress`

Note: If you have suggestions for improving this guide, open an issue or send a
pull request, as this is a topic under active development.
