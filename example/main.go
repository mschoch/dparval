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
