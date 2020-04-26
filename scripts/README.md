# Keyhole Cluster Information Collection script

## Command

```
mongo --quiet mongodb://localhost/ keyhole.js | tail -1 > cluster-info.json
```
