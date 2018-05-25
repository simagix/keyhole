# Seed Data
## Collections
### models

```
{
    "name" : "String(15)",
	"description" : "String(28)",
	"year" : 0
}
```

### robots

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

## Query Examples
### Find
```
db.getSisterDB("_KEYHOLE_").robots.find({ "_id" : "robot-1540a" })
```

### Range
```
db.getSisterDB("_KEYHOLE_").robots.find({ batteryPct: { $gt: .5} })
db.getSisterDB("_KEYHOLE_").robots.find({ batteryPct: { $lt: .5} })
db.getSisterDB("_KEYHOLE_").robots.find({ batteryPct: { $lt: .5, $gt: .2} })
```

### `$or`
```
db.getSisterDB("_KEYHOLE_").robots.find(
    { $or: [ {_id: "robot-1540a"}, {_id: "robot-22713"} ] }
)
```

### `$in` and `$nin`
```
db.getSisterDB("_KEYHOLE_").robots.find(
    { "_id" : { $in: ["robot-1540a", "robot-22713"] }}
)
    

db.getSisterDB("_KEYHOLE_").robots.find(
    { "_id" : { $nin: ["robot-1540a", "robot-22713"] }})
```

### `$elemMatch`
```
db.getSisterDB("_KEYHOLE_").robots.find(
    {tasks: {$elemMatch: {"minutesUsed": 25}} })
```

### `$lookup`
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