package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/gohxs/prettylog/global"
	"github.com/stdiopt/stream"
	"github.com/stdiopt/stream/format/strmjson"
	"github.com/stdiopt/stream/transform/strmagg"
	"github.com/stdiopt/stream/transport/strmhttp"
	"github.com/stdiopt/stream/utils/strmrefl"
	"github.com/stdiopt/stream/utils/strmutil"
)

// Experiment streaming values

func main() {
	l := stream.Line(
		strmutil.Value("https://randomuser.me/api/?results=100"),            // just sends the string
		strmhttp.GetFromInput(strmhttp.WithHeader("Authorization", "...}")), // fetches the url passed by the input
		strmjson.Decode(nil),        // Parses json if param is nil it will parse into an &interface{}
		strmrefl.Extract("results"), // from the map response, we send the contents of the field results
		strmrefl.Unslice(),          // unslice processes the incoming slice and send each element
		strmjson.Dump(os.Stdout),    // just writes json into stdout and passthru the elements
		// Aggregates the input elements and processes some fields
		// it produces a strmagg.Group that contains a slice of reduced fields
		strmagg.Aggregate(
			// manually processes the above message and returns the month of birth as the key for the group
			strmagg.GroupBy("birth months", func(v interface{}) (interface{}, error) {
				dob, ok := strmrefl.FieldOf(v, "dob", "date").(string)
				if !ok {
					return nil, errors.New("invalid field")
				}
				tm, err := time.Parse(time.RFC3339, dob)
				if err != nil {
					return nil, err
				}

				return tm.Month().String(), nil
			}),
			// Processes field "name" of the input and appends into a slice
			strmagg.Reduce("full name", func(a []string, v map[string]interface{}) []string {
				mv := v["name"].(map[string]interface{})
				return append(a, fmt.Sprintf("%s %s", mv["first"], mv["last"]))
			}),
			// Processes field "dob.age" from the input compares to the last returned one
			// and returns the maximum (older)
			strmagg.Reduce("older", func(a *float64, v interface{}) *float64 {
				f, ok := strmrefl.FieldOf(v, "dob", "age").(float64)
				if !ok {
					return a
				}
				if a == nil || f > *a {
					return &f
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
