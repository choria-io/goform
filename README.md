## Go Report Generator

This is a Go package for producing reports of row orientated data. It's inspired by [Perl Formats](https://perldoc.perl.org/perlform)
but not compatible with it.

Reports are pages of formatted rows with optional headers and footers per page. Summaries are automatically
kept for numeric data for use in headers and footers.

This project is not another template system like `text/template`, it's a report writer with no interactions 
for template functions and similar behaviors. Strings and Numbers are formatted using a `layout` format 
like `@<<<<<<<<<` for left justified text that does not exceed 10 characters. 

Today we support just textual output but PDF output could potentially be added in time.

## Status

This is quite a new project born mainly from missing the ease of use of Perl Formats in allowing
user supplied reports in CLI tooling.

It's actively being developed, and we're exploring expanding on the model a bit to give some 
more freedom via the ability to call functions to massage the data slightly.

## Example

Below we see some Go code to render a report, the header, body and footer can be supplied
either via code or in a YAML file using `NewFromFile`.

The intended outcome is that if given `[]any` data we can produce reports like the one below
based on layouts found in headers, body and footers.

```nohighlight
                                         Stream Report                                     Page 1 
--------------------------------------------------------------------------------------------------

ORDERS_0:                                              
    Messages: 1,000                    Size: 10 MiB                Consumers: 5                  
   First Seq: 3,198,803            Last Seq: 3,199,802              Subjects: 1                  

ORDERS_1:                                              
    Messages: 100                      Size: 10 MiB                Consumers: 5                  
   First Seq: 5,298,803            Last Seq: 5,299,803              Subjects: 1                  

--------------------------------------------------------------------------------------------------
     Streams: 2                    Subjects: 2                  
    Messages: 1,100                    Size: 20 MiB                Consumers: 10                      
--------------------------------------------------------------------------------------------------
```

### Header

Here we see a header printed top of each page, we have access to report and over all data but no row level data

Headers are optional, an empty string as header will disable it.

```nohighlight
      @||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||||| Page @#
      report.name,                                                                              report.page
--------------------------------------------------------------------------------------------------

```

### Body

In the body we tend to mainly access row data, each row is rendered using the body layout. Body layouts are required.

```nohighlight
@<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<:
row.config.name
Messages: @,#################       Size: @B#################   Consumers: @,#################
row.state.messages,                       row.state.bytes,                 row.state.consumer_count
First Seq: @,#################  Last Seq: @,#################    Subjects: @,#################
row.state.first_seq,                      row.state.last_seq,              row.state.num_subjects

```

### Footer

Footers are rendered at the end of the page, like headers they have no access to row data.

Here we see access to `report.summary.*`, any numeric field that was reported in the body layout
will automatically be accumulated here, rendering it like here will be a running total for the entire
report up to that point.

Footers are optional.

```nohighlight
--------------------------------------------------------------------------------------------------
Streams: @,#################  Subjects: @,#################
         report.current_row,            report.summary.state.num_subjects
Messages: @,#################      Size: @B#################   Consumers: @,#################     
          report.summary.state.messages, report.summary.state.bytes,      report.summary.state.consumer_count
--------------------------------------------------------------------------------------------------
```

### Calling

Here is basic go code to call the report setting a name and passing above header, body and footer
as parameters. 20 rows a page will be rendered with header/footer around each set.

```go
// header, footer, body are string vars holding the layout for those section
sr, _ := report.New("Stream Report", header, body, footer, 20)
sr.WriteReport(os.Stdout, data)
```

The report can be stored in a file - a YAML encoded `Report` - the report can then be generated
directly from the file:

```go
sr, _ := report.NewFromFile("report.yaml", "")
sr.WriteReport(os.Stdout, data)
```

We also support nested data where the rows are somewhere below the top level of data.

## Reference

### Nested Data

Let's say we want to render the `streams` item from this data:

```json
{
  "date": 1658413253,
  "server": "s.example.net",
  "results": {
    "streams": []
  }
}
```

We can render the nested data without first extracting it and access it in our report:

```go
sr, _ := report.NewFromFile("report.yaml","")
sr.WriteReportContainedRows(os.Stdout, data, "results.streams")
```

This will extract the `results.streams` for the individual `row` variables but the entire 
structure will be available in `data` in the report, here is a header showing the server
name from the initial data, this is useful to render data where report level information
is needed to create headers while individual rows are somewhere deeper.

```
@||||||||||||||||||||||||||||||| Page @#
data.server, report.page
----------------------------------------
```

### Available State

When listing the data to print like `report.name` or `row.state.bytes` we are 
writing [GJSON Path Queries](https://github.com/tidwall/gjson/blob/master/SYNTAX.md)
over a data structure, the structure looks like this:

```json
{
  "report": {
    "name": "Stream Report",
    "page": 1,
    "current_row": 10,
    "summary": {},
  },
  "row": {},
  "data": {},
}
```

| Item                         | Description                                                                                                 |
|------------------------------|-------------------------------------------------------------------------------------------------------------|
| `report.name`                | The name of the report as supplied in YAML or API calls                                                     |
| `report.page`                | The current page being rendered starting from 1                                                             |
| `report_current_row`         | The index of the current row being rendered, 0 based                                                        |
| `report.summary.state.bytes` | The accumulated sum of `row.state.bytes`, only tracks fields that was previously printed                    |
| `row.state.bytes`            | The `row` is the current row of data being rendered, `state.bytes` is a GJSON query into that structure     |
| `data.x`                     | When rendering rows from a nested structure based on a GJSON query, this is the entire original data source |

### Layouts

#### Strings

String data is formatted based on these layouts, strings are taken only up to the first line break,
if the length of a string exceeds that allowed by its layout the string is truncated.

| Format  | Description                                                                         |
|---------|-------------------------------------------------------------------------------------|
| `@>>>>` | Right justified string maximum 5 characters long                                    |
| `@<<<<` | Left justified string maximum 5 characters long                                     |
| `@||||` | Center justified string maximum 5 characters long                                   |
| `@>>>:` | Right justified string, maximum 5 characters long with a `:` suffix after the value |

#### Numbers

Numeric data is formatted on these layouts, if the resulting string is longer than the layout allows
a series of `#` characters will be returned indicating field overflow.

| Format      | Description                                                                                           |
|-------------|-------------------------------------------------------------------------------------------------------|
| `@########` | Renders a number with any gaps filled by space on the right                                           |
| `@,#######` | Renders a number in `comma` format, `1234` becomes `1,234`, space padded on the right                 |
| `@B#######` | Renders a number as base 1024 bytes, `1024` becomes `1KiB`, space padded on the right                 |
| `@.##`      | Renders a floating point number with specific precision, `1.234` becomes `1.12` but `10.12` overflows |
| `@#.##`     | Renders a floating point with specific precision, `10.234` becomes `10.12` but `100.123` overflows    |
| `@0.###`    | Renders a floating point with specific precision, padded by `0` on the left                           |
