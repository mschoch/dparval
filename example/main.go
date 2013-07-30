//  Copyright (c) 2013 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package main

import (
	"log"

	"github.com/mschoch/dparval"
)

func main() {
	// read some JSON off the wire
	bytes := []byte(`{"type":"test"}`)
	value := dparval.NewValueFromBytes(bytes)

	// is this actually JSON?
	if value.Type() == dparval.NOT_JSON {
		log.Printf("These bytes are not valid JSON")
	} else {

		// maybe the document had a field called "type"
		anotherVal, err := value.Path("type")

		if err != nil {
			err, ok := err.(*dparval.Undefined)
			if ok {
				log.Printf("type is undefined")
			} else {
				log.Printf("Unexpected error: %v", err)
			}
		} else {

			// anotherVal is another Value
			// we stay in this type system as long as possible
			// (so far we've avoided any fully JSON parsing)
			// lets see if type was a string
			if anotherVal.Type() == dparval.STRING {
				docType, err := anotherVal.Value()
				if err != nil {
					log.Printf("Unexpected error: %v", err)
				} else {
					log.Printf("The document type was %s", docType.(string))
				}
			}
		}
	}
}
