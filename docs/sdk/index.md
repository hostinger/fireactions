# Interacting with Fireactions via SDK

Fireactions provides a simple SDK for interacting with the Fireactions API. You can use the SDK to get, list, pause and resume pools.

Example usage:

```golang
package main

import (
  "fmt"
  "log"
  "os"

  "github.com/hostinger/fireactions"
)

func main() {
  client = fireactions.NewClient(fireactions.WithEndpoint(os.Getenv("FIREACTIONS_ENDPOINT")), fireactions.WithUsername(os.Getenv("FIREACTIONS_USERNAME")), fireactions.WithPassword(os.Getenv("FIREACTIONS_PASSWORD")))

  pool, err := client.GetPool("pool-id")
  if err != nil {
    log.Fatalf("Failed to get pool: %v", err)
  }

  fmt.Printf("Pool: %+v\n", pool)
}
```

For more up-to-date information, see the [GoDoc documentation](https://pkg.go.dev/github.com/hostinger/fireactions).
