---
description: |
    The checksum post-processor computes specified checksum for the artifact list
    from an upstream builder or post-processor. All downstream post-processors will
    see the new artifacts. The primary use-case is compute checksum for artifacts
    allows to verify it later.
layout: docs
page_title: 'Checksum Post-Processor'
...

# Checksum Post-Processor

Type: `checksum`

The checksum post-processor computes specified checksum for the artifact list
from an upstream builder or post-processor. All downstream post-processors will
see the new artifacts. The primary use-case is compute checksum for artifact to
verify it later.

After computes checksum for artifacts, you can use new artifacts with other
post-processors like
[artifice](https://www.packer.io/docs/post-processors/artifice.html),
[compress](https://www.packer.io/docs/post-processors/compress.html),
[docker-push](https://www.packer.io/docs/post-processors/docker-push.html),
[atlas](https://www.packer.io/docs/post-processors/atlas.html), or a third-party
post-processor.

## Basic example

The example below is fully functional.

``` {.javascript}
{
"type": "checksum"
}
```

## Configuration Reference

Optional parameters:

-   `types` (array of strings) - An array of strings of checksum types
to compute.
