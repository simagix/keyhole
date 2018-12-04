# Performance Comparison
The hypothesis is when searching on the first field of a document is faster than *n*th one.  But, will applying indexes make a difference?  This argument may be irrelevant because key-value pairs in a BSON document can have any order, except that _id is always first.  Let's find out from an experiment using *keyhole*.
 
## The Results
Without an index, searching on field *x* is slower than searching on *b*.  However, with index, `mongod` doesn't record logs granular enough to determine the winner.

```
+-------+--------+-------+-------+------+---------------------------------+-----------------------------------------------------------------------+
|Command|COLLSCAN| avg ms| max ms| Count| Namespace                       | Query Pattern                                                         |
|-------+--------+-------+-------+------+---------------------------------+-----------------------------------------------------------------------|
|find    COLLSCAN     456     751   2753 _KEYHOLE_88800.examples           {x:1}                                                                  |
|find    COLLSCAN     400     665   2753 _KEYHOLE_88800.examples           {b:1}                                                                  |
|find                   0      10   2753 _KEYHOLE_88800.examples           {a:1}                                                                  |
|find                   0      10   2753 _KEYHOLE_88800.examples           {w:1}                                                                  |
+-------+--------+-------+-------+------+---------------------------------+-----------------------------------------------------------------------+
```

## Test
### Spin up `mongod` 4.0 standalone
```
mlaunch init --dir ~/ws/analytic/data --single
```

### Schema Template
Create a *schema.json* template file with fields *a*, *b*, *w*, and *x* in email format for searching and other fields are fillers.

```
{
	"a": "ken.chen@mongodb.com",
	"b": "ken.chen@mongodb.com",
	"c": "I'm going to make him an offer he can't refuse.",
	"d": "Toto, I've a feeling we're not in Kansas anymore.",
	"e": "Here's looking at you, kid.",
	"f": "Go ahead, make my day.",
	"g": "All right, Mr. DeMille, I'm ready for my close-up.",
	"h": "May the Force be with you.",
	"i": "Fasten your seatbelts. It's going to be a bumpy night.",
	"j": "You talkin' to me?",
	"k": "What we've got here is failure to communicate.",
	"l": "I love the smell of napalm in the morning.",
	"m": "Love means never having to say you're sorry.",
	"n": "The stuff that dreams are made of.",
	"o": "E.T. phone home.",
	"p": "They call me Mister Tibbs!",
	"q": "You're gonna need a bigger boat.",
	"r": "Of all the gin joints in all the towns in all the world, she walks into mine.",
	"s": "Bond. James Bond.",
	"t": "There's no place like home.",
	"u": "Show me the money!",
	"v": "Frankly, my dear, I don't give a damn.",
	"w": "ken.chen@mongodb.com",
	"x": "ken.chen@mongodb.com"
}
```

### Seed 250,000 docs
*Keyhole* reads from the *schema.json* template and populates randomized data.

```
keyhole --uri mongodb://localhost/_KEYHOLE_88800 --seed \
    --file schema.json --total 250000
```

### Transactions Template
Create a *transactions.json* template file defining indexes on `{"a": 1}` and `{"w": 1}` and four queries below in a loop.  Queries on fields *b* or *x* will result collection scans.


```
{
	"indexes": [{
			"a": 1
		},
		{
			"w": 1
		}
	],
	"transactions": [{
			"c": "find",
			"filter": {
				"a": "ken.chen@mongodb.com"
			}
		},
		{
			"c": "find",
			"filter": {
				"w": "ken.chen@mongodb.com"
			}
		},
		{
			"c": "find",
			"filter": {
				"b": "ken.chen@mongodb.com"
			}
		},
		{
			"c": "find",
			"filter": {
				"x": "ken.chen@mongodb.com"
			}
		}
	]
}
```

### Set *slowms* to -1 capture all logs

```
db.setProfilingLevel(0, -1)
```

### Runs *Keyhole* in simulation mode
Runs *Keyhole* in simulation mode for 5 minutes with 10 concurrent connections and 20 TPS per connection.  *Keyhole* reads from *transactions.json* template and executes commands with randomized data.

```
keyhole --uri mongodb://localhost/_KEYHOLE_88800 --simonly \
    --tx transactions.json --tps 20 --conn 10
```

### Examine logs

```
keyhole --loginfo ~/ws/analytic/data/mongod.log
```

```
+-------+--------+-------+-------+------+---------------------------------+-----------------------------------------------------------------------+
|Command|COLLSCAN| avg ms| max ms| Count| Namespace                       | Query Pattern                                                         |
|-------+--------+-------+-------+------+---------------------------------+-----------------------------------------------------------------------|
|find    COLLSCAN     456     751   2753 _KEYHOLE_88800.examples           {x:1}                                                                  |
|find    COLLSCAN     400     665   2753 _KEYHOLE_88800.examples           {b:1}                                                                  |
|find                   0      10   2753 _KEYHOLE_88800.examples           {a:1}                                                                  |
|find                   0      10   2753 _KEYHOLE_88800.examples           {w:1}                                                                  |
+-------+--------+-------+-------+------+---------------------------------+-----------------------------------------------------------------------+
```