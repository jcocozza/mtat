# MTAT

MTA tool

## Subway Data

Subway routing and station data is provided by the [mta](https://www.mta.info/developers).
GTFS data can be found at `https://rrgtfsfeeds.s3.amazonaws.com/gtfs_subway.zip`.
This tools uses some, but not all of the data provided in the zip file.
For my use case(stops.txt), the data changes slowly.
As such, I've decided to just embed the data I need and refetch/recompile when necessary.
Use get_transit.sh to get the latest data.
