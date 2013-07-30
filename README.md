# dparval - a delayed parsing value system

The goal is offer a consistent type structure to allow you to process JSON end-to-end with minimal parsing.

Key aspects of the design:

* use unparsed JSON bytes to create Value object
* use Type() method to identify or respond to the type of data
* access nested values using Path() and Index() methods
    * these also return Value objects (or Undefined) and allow you to continue delaying parsing of nested objects
* SetPath() and SetIndex() allow you to overlay new values in objects and arrays
    * these do not disturb the original unparsed values
    * aliased values always take priority over original data in subsequent calls to Path() and Index()
    * aliased values are overlayed on top of the original if/when you call Value() to exit the type system
* call Value() at any time to exit the type system
    * this can be used to access the specific values you need (without ever parsing parts you didnt need)
    * this triggers actual parsing of the raw bytes
    * when aliases are used, additional objects may be created in this phase

Additional Features:

* attach arbitrary meta-data to any Value using AddMeta() methods
* read back stored meta-data using Meta()
* meta-data is just another Value

Future Ideas:

* Add ValueBytes() method which can go back to bytes
    * optimized to use raw bytes when no aliases were introduced

## example usage

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