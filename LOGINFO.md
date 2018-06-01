# Ops Performance Analytic
Display ops average execution with query patterns using `--loginfo` flag.

```
keyhole --uri mongodb://localhost/?replicaSet=replset --loginfo mongod.log
```

Below are sample outputs.

```
+-------+--------+-------+------+---------------------------------+----------------------------------------------------------------------+
|Command|COLLSCAN| avg ms| Count| Namespace                       | Query Pattern                                                        |
|-------+--------+-------+------+---------------------------------+----------------------------------------------------------------------|
|count  |COLLSCAN| 134277|     1|mongostore.user                  |{ query: { $or: [ { createdDate: { $gte: 1, $lt: 1 } }, {             |
|                                                                   deletedDate: { $gte: 1, $lt: 1 } } ] } }                             |
|delete |        |  64898|     1|mongostore.$cmd                  |{ deletes: [ { q: { created: { $lte: 1 } } } ] }                      |
|remove |        |  64897|     1|mongostore.taterCache            |{ created: { $lte: 1 } }                                              |
|find   |        |   9648|    19|mongostore.user                  |{ filter: { $and: [ { mongoweb.expirationDate: { $lte: 1 } }, {       |
|                                                                   mongoweb.statusNextCheck: { $exists: 1 } }, {                        |
|                                                                   mongoweb.statusNextCheck: { $lte: 1 } }, { mongoweb.cancelled: 1 } ] |
|                                                                   } }                                                                  |
|delete |        |    923|   287|mongostore.$cmd                  |{ deletes: [ { q: { tveUserId: 1 } } ], writeConcern: { w: 1 } }      |
|remove |        |    918|   289|mongostore.recentlyWatched       |{ tveUserId: 1 }                                                      |
|delete |        |    698|     1|mongostore.$cmd                  |{ deletes: [ { q: { device.deviceId: 1 } } ] }                        |
|remove |        |    680|     1|mongostore.activatingDevice      |{ device.deviceId: 1 }                                                |
|find   |        |    554|     2|mongostore.seriesSubscription    |{ filter: { userId: 1, seriesId: 1 } }                                |
|delete |        |    442|  2119|mongostore.$cmd                  |{ deletes: [ { q: { tveUserId: 1, titleId: { $in: [ ] } } } ],        |
|                                                                   writeConcern: { w: 1 } }                                             |
|remove |        |    441|  2119|mongostore.recentlyWatched       |{ tveUserId: 1, titleId: { $in: [ ] } }                               |
|find   |        |    401|     1|mongostore.user                  |{ filter: { deleted: 1, userProfile.firstnameLower: 1,                |
|                                                                   userProfile.lastnameLower: 1 } }                                     |
|find   |COLLSCAN|    399|     1|mongostore.faq                   |{ snapshot: 1 }                                                       |
|find   |        |    398|     2|mongostore.taterCache            |{ filter: { start: { $lt: 1 }, end: { $gt: 1 }, type: 1, displayType: |
|                                                                   1, platform: 1, feedVersion: 1 }, projection: { $sortKey: { $meta: 1 |
|                                                                   } }, sort: { created: -1 } }                                         |
|find   |        |    388|    47|mongostore.subscriptionTransactio|{ filter: { tveUserId: 1 }, projection: { $sortKey: { $meta: 1 } },   |
|                                                                   sort: { date: -1 } }                                                 |
|find   |        |    336|    43|mongostore.user                  |{ filter: { householdId: 1, deleted: { $ne: 1 } } }                   |
|find   |        |    322|     1|mongostore.featuredEntries       |{ projection: { $sortKey: { $meta: 1 } }, sort: { position: -1 } }    |
|find   |        |    303|     7|mongostore.activatingDevice      |{ filter: { device.deviceId: 1 }, projection: { $sortKey: { $meta: 1  |
|                                                                   } }, sort: { created: -1 } }                                         |
|find   |        |    281|     7|mongostore.schedule              |{ filter: { index: 1, displayType: 1, platform: 1, feedVersion: 1 } } |
|delete |        |    279|     2|mongostore.$cmd                  |{ deletes: [ { q: { tveUserId: 1 } } ] }                              |
|delete |        |    278|    24|mongostore.$cmd                  |{ deletes: [ { q: { userId: 1, seriesId: 1 } } ] }                    |
|remove |        |    278|    24|mongostore.seriesSubscription    |{ userId: 1, seriesId: 1 }                                            |
|find   |        |    276|    10|mongostore.manifestoResponse     |{ filter: { titleId: 1, adsAssetId: 1 }, projection: { $sortKey: {    |
|                                                                   $meta: 1 } }, sort: { timestampMillis: -1 } }                        |
|find   |        |    271|     1|mongostore.fairplayKey           |{ filter: { assetId: 1 } }                                            |
|delete |        |    266|   718|mongostore.$cmd                  |{ deletes: [ { q: { tveUserId: 1, titleId: 1 } } ], writeConcern: {   |
|                                                                   w: 1 } }                                                             |
|find   |        |    264|   364|mongostore.myList                |{ filter: { tveUserId: 1 }, projection: { $sortKey: { $meta: 1 } },   |
|                                                                   sort: { orderNumber: 1 } }                                           |
|update |        |    262|   288|mongostore.$cmd                  |{ updates: [ { q: { userId: 1, devices: { $elemMatch: { deviceId: 1 } |
|                                                                   } }                                                                  |
|update |        |    261|   288|mongostore.user                  |{ userId: 1, devices: { $elemMatch: { deviceId: 1 } } }               |
|remove |        |    260|   821|mongostore.recentlyWatched       |{ tveUserId: 1, titleId: 1 }                                          |
|find   |        |    255|  2989|mongostore.recentlyWatched       |{ filter: { tveUserId: 1, titleId: { $in: [ ] } }, projection: {      |
|                                                                   $sortKey: { $meta: 1 } }, sort: { updated: -1 } }                    |
|find   |        |    255|  1492|mongostore.recentlyWatched       |{ filter: { tveUserId: 1, titleId: 1 } }                              |
|find   |        |    254|  5578|mongostore.recentlyWatched       |{ filter: { tveUserId: 1 }, projection: { $sortKey: { $meta: 1 } },   |
|                                                                   sort: { updated: -1 } }                                              |
|remove |        |    253|     2|mongostore.user                  |{ _id: 1 }                                                            |
|delete |        |    253|     2|mongostore.$cmd                  |{ deletes: [ { q: { _id: 1 } } ] }                                    |
|update |        |    251|  3771|mongostore.user                  |{ userId: 1 }                                                         |
|update |        |    251|  3703|mongostore.$cmd                  |{ updates: [ { q: { userId: 1 }                                       |
|find   |        |    250|   266|mongostore.user                  |{ filter: { userNameLower: 1, deleted: 1 } }                          |
|update |        |    250|   449|mongostore.$cmd                  |{ updates: [ { q: { _id: 1 }                                          |
|update |        |    249|   450|mongostore.myList                |{ _id: 1 }                                                            |
|find   |        |    247|   899|mongostore.recentlyWatched       |{ filter: { tveUserId: 1, seriesId: 1 }, projection: { $sortKey: {    |
|                                                                   $meta: 1 } }, sort: { updated: -1 } }                                |
|update |        |    242|    15|mongostore.$cmd                  |{ updates: [ { q: { userId: 1, porku.statusLastChecked: { $lt: 1 } }  |
|update |        |    241|    15|mongostore.user                  |{ userId: 1, porku.statusLastChecked: { $lt: 1 } }                    |
|find   |        |    240|     1|mongostore.menu                  |{ filter: { displayType: 1, platform: 1, feedVersion: 1 } }           |
|find   |        |    239|   358|mongostore.user                  |{ filter: { userId: 1, deleted: { $ne: 1 } } }                        |
|find   |        |    239|   322|mongostore.user                  |{ filter: { devices: { $elemMatch: { ltlTokens.seriesId: 1 } },       |
|                                                                   deleted: { $ne: 1 } } }                                              |
|find   |        |    237|   577|mongostore.user                  |{ filter: { _id: 1 } }                                                |
|update |        |    235|  2705|mongostore.$cmd                  |{ updates: [ { q: { devices: { $elemMatch: { deviceId: 1,             |
|                                                                   deviceStatus: 1 } }, userId: { $ne: 1 } }                            |
|update |        |    235|  2705|mongostore.user                  |{ devices: { $elemMatch: { deviceId: 1, deviceStatus: 1 } }, userId:  |
|                                                                   { $ne: 1 } }                                                         |
|update |        |    233|  6526|mongostore.$cmd                  |{ updates: [ { q: { tveUserId: 1, titleId: 1 }                        |
|update |        |    233|  6526|mongostore.recentlyWatched       |{ tveUserId: 1, titleId: 1 }                                          |
|find   |        |    232|    63|mongostore.subscriptionTransactio|{ filter: { tveUserId: 1, transactionType: 1 } }                      |
|delete |        |    225|   103|mongostore.$cmd                  |{ deletes: [ { q: { tveUserId: 1, titleId: 1 } } ] }                  |
|find   |        |    220|   505|mongostore.user                  |{ filter: { ursRegId: 1, deleted: 1 } }                               |
|find   |        |    215|  2389|mongostore.resumeWatching        |{ filter: { userId: 1 } }                                             |
|update |        |    215|   164|mongostore.$cmd                  |{ updates: [ { q: { tveUserId: 1 }                                    |
|update |        |    215|   164|mongostore.biSubscriptionSummary |{ tveUserId: 1 }                                                      |
|find   |        |    207|   473|mongostore.user                  |{ filter: { msoUserId: 1, mso: 1 } }                                  |
|count  |        |    205|   485|mongostore.user                  |{ query: { msoHouseholdId: 1, mso: 1, deleted: 1 } }                  |
|update |        |    202|   298|mongostore.$cmd                  |{ updates: [ { q: { tveUserId: 1, devices: { $elemMatch: { deviceId:  |
|                                                                   1, ltlTokens.seriesId: 1 } } }                                       |
|find   |        |    202|   373|mongostore.myList                |{ filter: { tveUserId: 1, titleExpiry: { $gt: 1 } }, projection: {    |
|                                                                   $sortKey: { $meta: 1 } }, sort: { orderNumber: 1 } }                 |
|update |        |    202|   298|mongostore.user                  |{ tveUserId: 1, devices: { $elemMatch: { deviceId: 1,                 |
|                                                                   ltlTokens.seriesId: 1 } } }                                          |
|count  |        |    199|   304|mongostore.myList                |{ query: { tveUserId: 1, titleId: 1, titleExpiry: { $gt: 1 } } }      |
|find   |        |    190|    79|mongostore.user                  |{ filter: { porku.customerId: 1 } }                                   |
|update |        |    186|  1043|mongostore.$cmd                  |{ updates: [ { q: { tveUserId: 1, transactionType: 1, date: 1 }       |
|update |        |    186|  1043|mongostore.subscriptionTransactio|{ tveUserId: 1, transactionType: 1, date: 1 }                         |
|find   |        |    184|    38|mongostore.user                  |{ filter: { deleted: 1, userProfile.emailAddress: 1 } }               |
|find   |        |    183|   149|mongostore.user                  |{ filter: { hamdroid.purchaseToken: 1 } }                             |
|find   |        |    181|    14|mongostore.user                  |{ filter: { scrapple.originalTransactionId: 1 } }                     |
|find   |        |    178|     5|mongostore.homeFeed              |{ filter: { start: { $lt: 1 }, end: { $gt: 1 }, displayType: 1,       |
|                                                                   platform: 1, feedVersion: 1 }, projection: { $sortKey: { $meta: 1 }  |
|                                                                   }, sort: { created: -1 } }                                           |
|find   |COLLSCAN|    162|     3|mongostore.paywall               |{ filter: { startDate: { $lte: 1 }, endDate: { $gte: 1 } },           |
|                                                                   projection: { $sortKey: { $meta: 1 } }, sort: { weight: 1 } }        |
|find   |        |    158|     2|mongostore.castingAuth           |{ filter: { token: 1 } }                                              |
|find   |COLLSCAN|    147|     1|mongostore.dashLinearMediaInfo   |{ shardVersion: [ Timestamp 0|0, 1 ] }                                |
|find   |COLLSCAN|    140|     3|mongostore.series                |{ projection: { $sortKey: { $meta: 1 } }, sort: { name: 1 } }         |
|find   |        |    134|     1|mongostore.series                |{ filter: { seriesId: 1 } }                                           |
|find   |        |    134|     3|mongostore.user                  |{ filter: { msoHouseholdId: 1, accountType: 1 } }                     |
|find   |        |    134|     3|mongostore.user                  |{ filter: { mongoweb.amazonUserId: 1 } }                              |
|update |        |    132|     1|mongostore.$cmd                  |{ updates: [ { q: { tveUserId: 1, gatewayTransactionId: 1 }           |
|update |        |    132|     1|mongostore.subscriptionTransactio|{ tveUserId: 1, gatewayTransactionId: 1 }                             |
|find   |        |    132|    16|mongostore.user                  |{ filter: { $and: [ { hamdroid.statusNextCheck: { $exists: 1 } }, {   |
|                                                                   hamdroid.statusNextCheck: { $lte: 1 } }, { hamdroid.status: { $in: [ |
|                                                                   ] } } ] } }                                                          |
|find   |        |    131|    10|mongostore.user                  |{ filter: { userProfile.emailAddress: 1 } }                           |
|find   |        |    130|     1|mongostore.user                  |{ filter: { deleted: 1, userNameLower: 1 } }                          |
|remove |        |    130|     1|mongostore.resumeWatching        |{ userId: 1 }                                                         |
|delete |        |    130|     1|mongostore.$cmd                  |{ deletes: [ { q: { userId: 1 } } ], writeConcern: { w: 1 } }         |
|find   |        |    128|     8|mongostore.user                  |{ filter: { devices: { $elemMatch: { deviceId: 1, deviceStatus: 1 }   |
|                                                                   }, deleted: { $ne: 1 } } }                                           |
|find   |        |    112|     2|mongostore.user                  |{ filter: { $and: [ { mongoweb.expirationDate: { $gt: 1 } }, {        |
|                                                                   mongoweb.statusNextCheck: { $exists: 1 } }, {                        |
|                                                                   mongoweb.statusNextCheck: { $lt: 1 } } ] } }                         |
+-------+--------+-------+------+---------------------------------+----------------------------------------------------------------------+
```
