# Ops Performance Analytic
Display ops average execution with query patterns using `--loginfo` flag.

```
keyhole --uri mongodb://localhost/?replicaSet=replset --loginfo mongod.log
```

Below are sample outputs.

```
+-------+--------+-------+-------+------+---------------------------------+-----------------------------------------------------------------------+
|Command|COLLSCAN| avg ms| max ms| Count| Namespace                       | Query Pattern                                                         |
|-------+--------+-------+-------+------+---------------------------------+-----------------------------------------------------------------------|
|count   COLLSCAN  134277  134277      1 mongostore.user                   {$or: [ {createdDate: {$gte:1, $lt:1}}, {deletedDate: {$gte:1,         |
|                                                                            $lt:1}} ]}                                                           |
|delete             64898   64898      1 mongostore.$cmd                   {created: {$lte:1}}                                                    |
|remove             64897   64897      1 mongostore.taterCache             {created: {$lte:1}}                                                    |
|find                9648   12574     19 mongostore.user                   {$and: [ {spamazon.expirationDate: {$lte:1}},                          |
|                                                                            {spamazon.statusNextCheck: {$exists:1}}, {spamazon.statusNextCheck:  |
|                                                                            {$lte:1}}, {spamazon.cancelled:1} ]}                                 |
|remove               918    7936    289 mongostore.recentlyWatched        {tveUserId:1}                                                          |
|delete               698     698      1 mongostore.$cmd                   {device.deviceId:1}                                                    |
|remove               680     680      1 mongostore.activatingDevice       {device.deviceId:1}                                                    |
|find                 554     981      2 mongostore.seriesSubscription     {userId:1, seriesId:1}                                                 |
|remove               441    7631   2119 mongostore.recentlyWatched        {tveUserId:1, titleId: {$in: [...]}}                                   |
|find                 401     401      1 mongostore.user                   {deleted:1, userProfile.firstnameLower:1, userProfile.lastnameLower:1} |
|find                 398     480      2 mongostore.taterCache             {start: {$lt:1}, end: {$gt:1}, type:1, displayType:1, platform:1,      |
|                                                                            feedVersion:1}, sort: {created: -1}                                  |
|find                 388    2058     47 mongostore.subscriptionTransactio {tveUserId:1}, sort: {date: -1}                                        |
|find                 336    1841     43 mongostore.user                   {householdId:1, deleted: {$ne:1}}                                      |
|find                 322     322      1 mongostore.featuredEntries        {}, sort: {position: -1}                                               |
|find                 303     504      7 mongostore.activatingDevice       {device.deviceId:1}, sort: {created: -1}                               |
|find                 281     450      7 mongostore.schedule               {index:1, displayType:1, platform:1, feedVersion:1}                    |
|delete               279     286      2 mongostore.$cmd                   {tveUserId:1}                                                          |
|delete               278    1389     24 mongostore.$cmd                   {userId:1, seriesId:1}                                                 |
|remove               278    1389     24 mongostore.seriesSubscription     {userId:1, seriesId:1}                                                 |
|find                 276     470     10 mongostore.manifestoResponse      {titleId:1, adsAssetId:1}, sort: {timestampMillis: -1}                 |
|find    COLLSCAN     273     399      2 mongostore.dashLinearMediaInfo    {}                                                                     |
|find                 271     271      1 mongostore.fairplayKey            {assetId:1}                                                            |
|find                 264    2792    364 mongostore.myList                 {tveUserId:1}, sort: {orderNumber:1}                                   |
|update               261    1332    576 mongostore.$cmd                   {userId:1, devices: {$elemMatch: {deviceId:1}}}                        |
|remove               260    2995    821 mongostore.recentlyWatched        {tveUserId:1, titleId:1}                                               |
|find                 255    2594   2989 mongostore.recentlyWatched        {tveUserId:1, titleId: {$in: [...]}}, sort: {updated: -1}              |
|find                 255    2087   1492 mongostore.recentlyWatched        {tveUserId:1, titleId:1}                                               |
|find                 254    3661   5578 mongostore.recentlyWatched        {tveUserId:1}, sort: {updated: -1}                                     |
|delete               253     275      2 mongostore.$cmd                   {_id:1}                                                                |
|remove               253     275      2 mongostore.user                   {_id:1}                                                                |
|update               251    2090   7474 mongostore.$cmd                   {userId:1}                                                             |
|find                 250    2041    266 mongostore.user                   {userNameLower:1, deleted:1}                                           |
|update               250    2171    899 mongostore.$cmd                   {_id:1}                                                                |
|find                 247    4150    899 mongostore.recentlyWatched        {tveUserId:1, seriesId:1}, sort: {updated: -1}                         |
|update               242    1155     30 mongostore.$cmd                   {userId:1, porku.statusLastChecked: {$lt:1}}                           |
|find                 240     240      1 mongostore.menu                   {displayType:1, platform:1, feedVersion:1}                             |
|find                 239     525    358 mongostore.user                   {userId:1, deleted: {$ne:1}}                                           |
|find                 239    2039    322 mongostore.user                   {devices: {$elemMatch: {ltlTokens.seriesId:1}}, deleted: {$ne:1}}      |
|find                 237    2084    577 mongostore.user                   {_id:1}                                                                |
|update               235    2109   5410 mongostore.$cmd                   {devices: {$elemMatch: {deviceId:1, deviceStatus:1}}, userId: {$ne:1}} |
|update               233    3609  13052 mongostore.$cmd                   {tveUserId:1, titleId:1}                                               |
|find                 232    1443     63 mongostore.subscriptionTransactio {tveUserId:1, transactionType:1}                                       |
|delete               225    1568    103 mongostore.$cmd                   {tveUserId:1, titleId:1}                                               |
|find                 220    2076    505 mongostore.user                   {ursRegId:1, deleted:1}                                                |
|find                 215    2085   2389 mongostore.resumeWatching         {userId:1}                                                             |
|update               215    1914    328 mongostore.$cmd                   {tveUserId:1}                                                          |
|find                 207    2069    473 mongostore.user                   {msoUserId:1, mso:1}                                                   |
|count                205    1964    485 mongostore.user                   {msoHouseholdId:1, mso:1, deleted:1}                                   |
|find                 202    2048    373 mongostore.myList                 {tveUserId:1, titleExpiry: {$gt:1}}, sort: {orderNumber:1}             |
|update               202    2077    596 mongostore.$cmd                   {tveUserId:1, devices: {$elemMatch: {deviceId:1,                       |
|                                                                            ltlTokens.seriesId:1}}}                                              |
|count                199    2035    304 mongostore.myList                 {tveUserId:1, titleId:1, titleExpiry: {$gt:1}}                         |
|find                 190     925     79 mongostore.user                   {porku.customerId:1}                                                   |
|update               186    2051   2086 mongostore.$cmd                   {tveUserId:1, transactionType:1, date:1}                               |
|find                 184     996     38 mongostore.user                   {deleted:1, userProfile.emailAddress:1}                                |
|find                 183    2081    149 mongostore.user                   {hamdroid.purchaseToken:1}                                             |
|find                 181     630     14 mongostore.user                   {scrapple.originalTransactionId:1}                                     |
|find                 178     269      5 mongostore.homeFeed               {start: {$lt:1}, end: {$gt:1}, displayType:1, platform:1,              |
|                                                                            feedVersion:1}, sort: {created: -1}                                  |
|find    COLLSCAN     162     238      3 mongostore.paywall                {startDate: {$lte:1}, endDate: {$gte:1}}, sort: {weight:1}             |
|find                 158     212      2 mongostore.castingAuth            {token:1}                                                              |
|find    COLLSCAN     140     142      3 mongostore.series                 {}, sort: {name:1}                                                     |
|find                 134     134      1 mongostore.series                 {seriesId:1}                                                           |
|find                 134     138      3 mongostore.user                   {msoHouseholdId:1, accountType:1}                                      |
|find                 134     138      3 mongostore.user                   {spamazon.amazonUserId:1}                                              |
|update               132     132      2 mongostore.$cmd                   {tveUserId:1, gatewayTransactionId:1}                                  |
|find                 132     264     16 mongostore.user                   {$and: [ {hamdroid.statusNextCheck: {$exists:1}},                      |
|                                                                            {hamdroid.statusNextCheck: {$lte:1}}, {hamdroid.status: {$in:        |
|                                                                            [...]}} ]}                                                           |
|find                 131     133     10 mongostore.user                   {userProfile.emailAddress:1}                                           |
|remove               130     130      1 mongostore.resumeWatching         {userId:1}                                                             |
|find                 130     130      1 mongostore.user                   {deleted:1, userNameLower:1}                                           |
|find                 128     138      8 mongostore.user                   {devices: {$elemMatch: {deviceId:1, deviceStatus:1}}, deleted:         |
|                                                                            {$ne:1}}                                                             |
|find                 112     113      2 mongostore.user                   {$and: [ {spamazon.expirationDate: {$gt:1}},                           |
|                                                                            {spamazon.statusNextCheck: {$exists:1}}, {spamazon.statusNextCheck:  |
|                                                                            {$lt:1}} ]}                                                          |
+-------+--------+-------+-------+------+---------------------------------+-----------------------------------------------------------------------+
```
