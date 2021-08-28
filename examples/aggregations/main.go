package main

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/gohxs/prettylog/global"
	"github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmagg"
	"github.com/stdiopt/stream/strmhttp"
	"github.com/stdiopt/stream/strmjson"
	"github.com/stdiopt/stream/strmrefl"
	"github.com/stdiopt/stream/strmutil"
)

// Experiment streaming values

func main() {
	l := stream.Line(
		strmutil.Value("https://randomuser.me/api/?results=100"),   // just sends the string
		strmhttp.Get(strmhttp.WithHeader("Authorization", "...}")), // fetches the url passed by the input
		strmjson.Decode(nil),        // Parses json if param is nil it will parse into an &interface{}
		strmrefl.Extract("results"), // from the map response, we send the contents of the field results
		strmrefl.Unslice(),          // unslice processes the incoming slice and send each element
		strmjson.Dump(os.Stdout),    // just writes json into stdout and passthru the elements
		// Aggregates the input elements and processes some fields
		// it produces a strmagg.Group that contains a slice of reduced fields
		strmagg.Aggregate(
			// manually processes the above message and returns the month of birth as the key for the group
			strmagg.GroupBy("birth months", func(v interface{}) (interface{}, error) {
				dob, err := strmrefl.FieldOf(v, "dob", "date")
				if err != nil {
					return nil, err
				}
				tm, err := time.Parse(time.RFC3339, dob.(string))
				if err != nil {
					return nil, err
				}

				return tm.Month().String(), nil
			}),
			// Processes field "name" of the input and appends into a slice
			strmagg.Reduce("full name", strmagg.Field("name"), func(a []string, v map[string]interface{}) []string {
				return append(a, fmt.Sprintf("%s %s", v["first"], v["last"]))
			}),
			// Processes field "dob.age" from the input compares to the last returned one
			// and returns the maximum (older)
			strmagg.Reduce("older", strmagg.Field("dob", "age"), func(a *float64, v float64) *float64 {
				if a == nil || v > *a {
					return &v
				}
				return a
			}),
		),
		// write json of the aggregation to stdout
		strmjson.Dump(os.Stdout),
	)
	err := stream.Run(l)
	if err != nil {
		log.Println("err:", err)
	}
}
