# mongo

## How to

```bash
go get github.com/tel-io/instrumentation/plugins/mongo@latest
```

### Usage
```go
package main

import (
	"context"
	"github.com/tel-io/tel/v2"
	plugin "github.com/tel-io/instrumentation/plugins/mongo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	t, cc := tel.New(context.Background(), tel.GetConfigFromEnv())
	defer cc()
	
	// connect to MongoDB
	opts := options.Client()

	// inject plugin
	plugin.Inject(opts, plugin.WithTel(&t))

	opts.ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), opts)

}
```