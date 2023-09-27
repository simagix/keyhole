# Keyhole - Survey Your Mongo Land

Keyhole is a tool to explore MongoDB deployments. Instructions are available at

[![Survey Your Mongo Land with Keyhole and Maobi](https://img.youtube.com/vi/kObLsYJAruI/0.jpg)](https://youtu.be/kObLsYJAruI?si=Tv2Qbd2vHATt0WH1).

For MongoDB JSON logs parsing, use the improved [Hatchet](https://github.com/simagix/hatchet) project.

## License

[Apache-2.0 License](LICENSE)

## Disclaimer

This software is not supported by MongoDB, Inc. under any of their commercial support subscriptions or otherwise. Any usage of keyhole is at your own risk.

## Changes
### v1.3.x
- `-allinfo` supports high number of collections
- `-loginfo` includes raw logs

### v1.2.1
- Supports ReadPreferenceTagSets

### v1.2
- Prints connected client information from all `mongod` processes
- Prints client information from a log file
- Compares two clusters using internal metadata
- Performs deep comparison of two clusters
