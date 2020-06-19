# easyjson-unixtimestamp

## WHY?
Package [easyjson](https://github.com/mailru/easyjson) provides a fast and easy way to marshal/unmarshal Go structs to/from JSON without the use of reflection. 

But what if your JSON contains time as unix epoch in seconds and in your Go struct you want to use `time.Time`?

```
type SensorData struct {
	SensorType SensorType `json:"type"`
	SensorSN   string     `json:"sensor_sn"`
	Timestamp  time.Time  `json:"timestamp"`
	Value      float64    `json:"value"`
}
```

```
{
    "type": "tmp",
    "sensor_sn": "00000000-0000-0000-0000-000000000000",
    "timestamp": 1592577254,
    "value": 13
}
```

## HOW?
There are many ways to achieve this:
- Write custom MarshalEasyJSON/UnmarshalEasyJSON
- Create struct method to return time.Time
- Use different tools like `grep`, `sed` or `awk` to change your easyjson generated go file.
- Use this utility 

## WHAT?

This utility will parse easyjson generated file into AST and modify the encoder and decoder to do `timestamp <-> time.Time` transformations

```
go get github.com/troyanov/easyjson-unixtimestamp
```

```
Usage of easyjson-unixtimestamp
    -file string
        path to generated _easyjson.go file
    -jsonTag string
        json tag for the timestamp field that is unix timestamp (default "timestamp")
    -structField string
        struct field name that contains timestamp as time.Time (default "Timestamp")
```
