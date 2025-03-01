This project is licensed under the MIT License. See the [LICENSE](./LICENSE) file for more details.

# JTL

JTL (stands for junior tag language) is a simple tag language with a "nicer" syntax.

example
```jtl
DOCTYPE=JTL

>>>ENV;
    >>>NAME=me
>>>version=1.0;

>>>BEGIN;
    >//>this is a comment
    >type="content/lua">script>"
local content = "$env:NAME";
";
<<<ENDL
>>>END;
```
would do
```go
map[string]interface{}{
    "script": map[string]interface{}{
        "type": "content/lua",
        "Content": `local content = "me"`
    }
}
```

I made this because html and xml were annoying to write all that in. I call it productivity.

import:

```go
import (
    // other imports...
    "fmt"
    "encoding/json"
    "github.com/OrtheSnowJames/jtl"
)
```

tidy
```sh
go mod tidy
```

parse
```go
func main() {
    parsed, err := jtl.Parse(`
>>>DOCTYPE=JTL

>>>ENV;
    >>>NAME=me
>>>version=1.0;

>>>BEGIN;
    >//>this is a comment
    >type="lua">script>
        document.onEvent(".buttontest", "click", [[
            print("Button clicked!")
            -- Do more stuff here
        ]]);
";
<<<ENDL
>>>END;
    `)
    if err != nil {
        fmt.Printf("error parsing: %s \n", err)
        return
    } else {
        bytedat, err := json.Marshall(parsed)
        if err != nil {
            fmt.Printf("error marshalling json: %s \n", err)
            return
        }
        fmt.Println(string(bytedat))
    }
}
```

And there you have it. You sucessfully parsed a jtl document and printed it. What you do with the data now, no one knows.

yes, tests were passed (2/29/2025)
```output
Running tool: /usr/bin/go test -timeout 30s -coverprofile=/tmp/vscode-go1krrVp/go-code-cover github.com/OrtheSnowJames/jtl

ok  	github.com/OrtheSnowJames/jtl	0.004s	coverage: 95.1% of statements
```

Latest Commit: Allows you to make blank tags
You can now do this
```
>>>DOCTYPE=JTL;

>>>BEGIN;
    >src="test.lua">script>;
    >class="empty">div>;
>>>END;
```
.