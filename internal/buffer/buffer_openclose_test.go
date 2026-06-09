package buffer

import (
	"testing"

	"github.com/micro-editor/micro/v2/internal/config"
	ulua "github.com/micro-editor/micro/v2/internal/lua"
	lua "github.com/yuin/gopher-lua"
)

func init() {
	ulua.L = lua.NewState()
	config.InitRuntimeFiles(false)
	config.InitGlobalSettings()
	config.GlobalSettings["backup"] = false
	config.GlobalSettings["fastdirty"] = true
}

func TestBufferOpenClosePanic(t *testing.T) {
	NewBufferFromString("a", "", BTDefault)
	NewBufferFromString("b", "", BTDefault)
	CloseOpenBuffers()
}
