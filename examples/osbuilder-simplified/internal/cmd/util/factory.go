package util

import (
	"fmt"
	"os"

	"github.com/onexstack/osbuilder/examples/osbuilder-simplified/internal/util/options"
)

// Factory 接口 —— 严格对应 util/factory_client_access.go 的 Factory
type Factory interface {
	GetOptions() *options.Options
}

type factoryImpl struct{ opts *options.Options }

func (f *factoryImpl) GetOptions() *options.Options { return f.opts }

func NewFactory(opts *options.Options) Factory { return &factoryImpl{opts: opts} }

// CheckErr 对应 cmdutil.CheckErr（Run 中串联 Complete/Validate/Run 的统一错误处理）
func CheckErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
