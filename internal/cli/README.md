# Pureinsights Discovery Platform: Command Line Interface

## Printer
Printer is a type definition for the functions that print the JSON objects used in Discovery:
```go
type Printer func(iostreams.IOStreams, ...gjson.Result) error
```

Various printers can be written to change the format in which a JSON object is displayed to the user.

The current printers are the following:
| Name | Format | Options | Description | Usage |
| --- | --- | --- | --- | --- |
| JsonObjectPrinter | JSON | Pretty | Prints a single JSON object in JSON format. If the `pretty` boolean is true, then the JSON is printed with spacing and indentation. If not, it is printed in a compact format. |  ``` jsonPrinter := JsonObjectPrinter(pretty) ``` <br> ``` jsonPrinter(iostreams, json) ``` |
| JsonArrayPrinter | JSON Array | Pretty | Prints a JSON array in JSON format. If the `pretty` boolean is true, then the elements in the array are pretty-printed and brackets (`[]`) are included at each end. If not, each of the elements is printed on a single line. |  ``` arrayPrinter := JsonArrayPrinter(pretty) ``` <br> ``` arrayPrinter(iostreams, array...) ``` |


### Register a new printer
To add a new printer, the first step is to add the print function in the `printer.go` file. This function should transform the received JSON into the desired format and define the different display options.

Then, a constructor for that format must be created. This function returns a Printer, so it returns a function that receives `IOStreams` and `...gjson.Result`, prints the JSON object in the new format, and returns an error if any occured.

Finally, the format must be added to the `GetObjectPrinter(name)` or `GetArrayPrinter(name)` functions. A new case must be added to the switch statement. The case returns the constructor of the format. After this, the printer was successfully registered and can be used.