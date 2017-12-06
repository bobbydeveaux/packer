package main

import (
	"github.com/mitchellh/packer/builder/vcloud"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(new(vcloud.Builder))
	server.Serve()
}
