# Keyhole Cluster Information Collection script

## Command

```
mongo --quiet mongodb://localhost/ keyhole.js | grep -v '^202' > cluster-info.json
```
