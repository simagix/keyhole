# Cluster Info and Stats

Keyhole with `--allinfo` flag collects cluster and storage information:

- Sharded cluster
- Replica set
- Standalone

## Usages

```bash
keyhole --allinfo <connection_string> [--obfuscate]
```

## Cluster Information Collection

With `--allinfo`, it collects cluster stats and output to a gzipped BSON file.  You can use [Maobi](https://hub.docker.com/repository/docker/simagix/maobi) to generate a report. For example:

```bash
keyhole --allinfo "mongodb://..."
```

For a sharded cluster, Keyhole collects chunks information to create Shard Distribution information.  Note that, with thousands of chunks, collecting chunk sizes is a time consuming process.

## Obfuscation

To avoid leaking PII and PHI information, you can obfuscate the sample document collected by *Keyhole*.  For example:

```bash
keyhole --allinfo "mongodb://..." --obfuscate
```

**Note:** The `--redact` flag is also supported and functions identically to `--obfuscate`.

**Note:** The connection string must immediately follow `--allinfo`, or use `=` syntax:

```bash
keyhole --allinfo="mongodb://..." --obfuscate
```