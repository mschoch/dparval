# dparval - a delayed parsing value system

The goal is offer a consistent type structure to allow you to process JSON end-to-end with minimal parsing.

### Features

* Create Value objects with unparsed JSON bytes or with JSON compatible go datatypes
* Access nested data using Path() and Index() methods
    * these also return Value objects (or Undefined) allowing you to delay parsing of nested objects as well
* Values of type OBJECT and ARRAY are mutable
    * SetPath() and SetIndex() allow you to overlay new Values into these objects
* Returning to []byte can be done at any time by calling Bytes()
    * the underlying raw bytes are reused if possible to avoid JSON encoding
* Exit the type system at any time by calling Value()
    * this will trigger parsing of any required values that have not yet been parsed
* Arbitrary data may be attached to a Value using the Set/Get/Remove Attachment() methods
* Check the type of Value using the Type() method

### Documentation

See [GoDoc](http://godoc.org/github.com/mschoch/dparval)

### Performance

Two simple benchmarks which process a 1MB JSON file.  The first uses NewValueFromBytes() with Path(), Index() and Value() calls and the second using json.Unmarshal() and map/slice calls.  Both versions access the same property.

    $ go test -bench .
    PASS
    BenchmarkLargeValue	      20	  76905386 ns/op	  25.23 MB/s
    BenchmarkLargeMap	      20	 116378934 ns/op	  16.67 MB/s
    ok  	github.com/mschoch/dparval	4.179s

### Usage

	// read some JSON
	bytes := []byte(`{"type":"test"}`)

	// create a Value object
	doc := dparval.NewValueFromBytes(bytes)

	// attempt to access a nested Value
	docType, err := doc.Path("type")
	if err != nil {
		panic("no property type exists")
	}

	// convert docType to a native go value
	docTypeValue := docType.Value()

	// display the value
	fmt.Printf("document type is %v\n", docTypeValue)

### Output

    $ ./example
    document type is test
