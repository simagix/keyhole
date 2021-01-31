# Seed Data
keyhole can read a sample document and generate dummy data.  Here is an example.

## Usages

```
keyhole -seed [--drop] [--total <num>] [--file <template>] <connection_string>
```

## Create a Template
```
keyhole -v --schema --collection favorites mongodb://localhost/keyhole > /tmp/template.json
```

## Seed Data
```
keyhole --seed --total 10000 --file /tmp/template.json --collection xyz mongodb://localhost/keyhole
```