package main

import (
  "context"

  "github.com/benthosdev/benthos/v4/public/service"

  _ "github.com/benthosdev/benthos/v4/public/components/all"

  _ "github.com/vuldin/benthos-decrypt/processor"
)

func main() {
  service.RunCLI(context.Background())
}
