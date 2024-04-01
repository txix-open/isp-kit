package rc

import (
	"github.com/txix-open/isp-kit/rc/schema"
)

func GenerateConfigSchema(cfgPtr any) schema.Schema {
	s := schema.NewGenerator().Generate(cfgPtr)
	s.Title = "Remote config"
	s.Version = ""
	return s
}
