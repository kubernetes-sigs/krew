---
title: Plugin Naming Guide
slug: naming-guide
weight: 200
---

This document explains the best practices and recommendations for naming your
kubectl plugins.

These guidelines are used for reviewing the plugins [submitted to Krew]({{<ref
"../release/submitting-to-krew.md">}}).

##### _Punctuation_

Plugin names must be all lowercase and separate words with hyphens.
**Don't** use camelCase, PascalCase, or snake_case; use
[kebab-case](http://wiki.c2.com/?KebabCase).

- **DON'T:** `kubectl OpenSvc`
- **DO:** `kubectl open-svc`

##### _Be specific_

Plugin names should not be verbs/nouns that are generic, already overloaded, or
possibly can be used for broader purposes by another plugin.

- **DON'T:** `kubectl login`: Tries to put dibs on the word.
- **DO:** `kubectl gke-login`.

Also:

- **DON'T:** `kubectl ui`: Should be used only for Kubernetes Dashboard.
- **DO:** `kubectl gke-ui`.

##### _Be unique_

Try to find a unique name for your plugin that differentiates you from other
possible plugins doing the same job.

- **DON'T:** `kubectl view-logs`: Unclear how it is different than the builtin
  "logs" command, or many other tools for viewing logs.
- **DO:** `kubectl tailer`:  Unique name, points to the underlying
  tool name.

##### _Use Verbs/Resource Types_

If the name does not make it clear (a) what verb the plugin is doing on a
resource, or (b) what kind of resource it's doing the action on, consider
clarifying unless it is obvious.

- **DON'T:** `kubectl service`: Unclear what this plugin is doing with
  service.
- **DON'T:** `kubectl open`: Unclear what it is opening.
- **DO:** `kubectl open-svc`: It is clear the plugin will open a service.

##### _Prefix Vendor Identifiers_

Use the vendor-specific strings as prefix, separated with a dash. This makes it
easier to search/group plugins that are about a specific vendor.

- **DON'T:** `kubectl ui-gke`: Makes it harder to search or locate in a
  plugin list.
- **DO:** `kubectl gke-ui`: Will show up next to other gke-* plugins.

##### _Avoid repeating kube[rnetes]_

Plugin names should not repeat kube- or kubernetes- prefixes to avoid
stuttering.

- **DON'T:** `kubectl kube-node-admin`: "kubectl " already has "kube" in
  it.
- **DO:** `kubectl node-admin`.

##### _Avoid Resource Acronyms_

Using kubectl acronyms for API resources (e.g. svc, ing, deploy, cm) reduces
readability and discoverability of a plugin more than it is saving keystrokes.

- **DON'T:** `kubectl new-ing`: Hard to spot and the plugin is for
  Ingress.
- **DO:** `kubectl debug-ingress`.

If you have suggestions to this guide, open an issue or send a pull request, as
this is an open topic of debate with a lot of gray areas.
