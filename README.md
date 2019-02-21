# mkclient

A package that generates a Go client from a [Swagger schema](github.com/marwan-at-work/swag)

### Usage

```golang

import (
    "marwan.io/swag"
    "marwan.io/mkrest"
)

func main() {
    api := swag.New(
        // all API endpoints
    )
    err := mkrest.MakeClient(api)
}
```

The above line will generate a client package for you based on all swagger endpoints. 
Each endpoint will take its arguments from the defined query strings and body request. 