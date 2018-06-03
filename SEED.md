# Seed Data
## Seed data from template
Seed data by reading a template from a file.  Key characters of random values are

- Random email addresses
- Random IP addresses
- Random date strings
- Random HEX string
- Random string having the same length of the original string
- Radnom numbers

```
keyhole -uri=mongodb://localhost/?replicaSet=replset -seed --file file.json [--drop]
```


### Template
```
{
	"email": "simagix@gmail.com",
	"emailAddr": "ken.chen@mongodb.com",
	"hostIP": "192.168.1.1",
	"string1": "I have a string value",
	"string2": "This is another string value. You can use any field names",
	"number1": 123,
	"number2": 456,
	"hex1": "12345678",
	"hex2": "a1023b435c893d123e567f3487",
	"lastUpdated": {"$date": "2018-01-01T01:23:45Z"}
}
```

### A Result Example
```
{
	"_id" : ObjectId("5b1355014a24e39842b78656"),
	"email" : "James.Miller@simagix.com",
	"emailAddr" : "Liam.Garcia@outlook.com",
	"hostIP" : "11.72.211.45",
	"string1" : "Toto, I've a",
	"string2" : "The first rule of Fight Club is: You do not talk about",
	"number1" : 4059,
	"number2" : 563,
	"hex1" : "dcd554c2",
	"hex2" : "24a621e2cede763b679dc88d11",
	"lastUpdated" : {
		"$date" : "2011-02-21T04:52:24-05:00"
	}
}
```

## Seed data for demo.

```
keyhole -uri=mongodb://localhost/?replicaSet=replset -seed
```

### Collections
#### models

```
{
    "name" : "String(15)",
	"description" : "String(28)",
	"year" : 0
}
```

#### robots

```
{
	"modelId" : "String(12)",
	"notes" : "String(20)",
	"batteryPct" : 0,
	"tasks" : [
		{
			"for" : "String(8)",
			"minutesUsed" : 0
		},
		{
			"for" : "String(4)",
			"minutesUsed" : 0
		}
	]
}
```

### Query Examples
#### Find
```
db.getSisterDB("_KEYHOLE_").robots.find({ "_id" : "robot-1540a" })
```

#### Range
```
db.getSisterDB("_KEYHOLE_").robots.find({ batteryPct: { $gt: .5} })
db.getSisterDB("_KEYHOLE_").robots.find({ batteryPct: { $lt: .5} })
db.getSisterDB("_KEYHOLE_").robots.find({ batteryPct: { $lt: .5, $gt: .2} })
```

#### `$or`
```
db.getSisterDB("_KEYHOLE_").robots.find(
    { $or: [ {_id: "robot-1540a"}, {_id: "robot-22713"} ] }
)
```

#### `$in` and `$nin`
```
db.getSisterDB("_KEYHOLE_").robots.find(
    { "_id" : { $in: ["robot-1540a", "robot-22713"] }}
)
    

db.getSisterDB("_KEYHOLE_").robots.find(
    { "_id" : { $nin: ["robot-1540a", "robot-22713"] }})
```

#### `$elemMatch`
```
db.getSisterDB("_KEYHOLE_").robots.find(
    {tasks: {$elemMatch: {"minutesUsed": 25}} })
```

#### `$lookup`
```
db.getSisterDB("_KEYHOLE_").robots.aggregate([
    {
        "$lookup": {
            "from": "models",
            "localField": "modelId",
            "foreignField": "_id",
            "as": "meta"
        }
    },
    {
        "$unwind": {
            "path": "$meta",
            "preserveNullAndEmptyArrays": true
        }
    },
    {
        "$project": {
            "_id": "$_id",
            "description": "$meta.description",
            "modelId": "$modelId",
            "batteryPct": "$batteryPct",
            "notes": "$notes"
        }
    }
])
```