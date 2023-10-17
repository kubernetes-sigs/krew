---
title: Testing plugin installation locally
slug: testing-locally
weight: 300
---

After you have written your [plugin manifest]({{< ref "plugin-manifest.md" >}})
and archived your plugin into a `.zip` or `.tar.gz` file, you can verify that
your plugin installs correctly with Krew by running:

```sh
{{<prompt>}}kubectl krew install --manifest=foo.yaml --archive=foo.tar.gz
```

- The `--manifest` flag specifies a custom manifest rather than using
  the default [krew index][index]
- `--archive` overrides the download `uri:` specified in the plugin manifest and
  uses a local `.zip` or `.tar.gz` file instead.

If the installation **fails**, run the command again with `-v=4` flag to see the
verbose logs and examine what went wrong.

If the installation **succeeds**, you should now be able to run your plugin.

If you made your archive file available for download on the Internet, run the
same command without the `--archive` option and actually test downloading the
file from the specified `uri` and validate its `sha256` sum is correct.

After you have tested your plugin installation, uninstall it with `kubectl krew uninstall foo`.

### Testing other platforms

If you need to test other `platforms` definitions that don't match your current machine,
you can use the `KREW_OS` and `KREW_ARCH` environment variables to override the
OS and architecture that Krew thinks it's running on.

For example, if you're on a Linux machine, you can test Windows installation
with:

```sh
{{<prompt>}}KREW_OS=windows KREW_ARCH=amd64 kubectl krew install --manifest=[...]
```

[index]: https://github.com/kubernetes-sigs/krew-index
