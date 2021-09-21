# stream

**API is not stable and might change eventually**

An experiment around building composable, streamable pipelines in go

the idea it self is not about maximum performance but rather an simplier way to
abstract channels, concurrency, cancellation and dynamic data mostly for
ETL jobs. It relies heavily on `interface{}` to pass data around and has some
heavy reflection usage on `github.com/stdiopt/stream/strmutil`.

it's be possible to build procs around serializing CSVs, querying RDBs, producing
Parquet, crawling url's, consuming API's, etc...

## Usage example

This example will fetch a zip file and produce a parquet from a csv file inside the
zip

```go
func main() {
	err := strm.Run(
		strmhttp.Get("https://datahub.io/core/world-cities/r/world-cities_zip.zip"),
		strmzip.Stream("world-cities_csv.csv"),
		strmcsv.DecodeAsDrow(','),
		strmdrow.NormalizeColumns(),
		strmparquet.Encode()
		strmfs.WriteFile("cities.parquet")
	)
	if err != nil {
		log.Fatal(err)
	}
}
```
