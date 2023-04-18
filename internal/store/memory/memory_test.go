package memory_test

import (
	"testing"

	"github.com/Fzgt/cronflux/internal/store"
	"github.com/Fzgt/cronflux/internal/store/memory"
	"github.com/Fzgt/cronflux/internal/store/storetest"
)

func TestMemoryStoreConformance(t *testing.T) {
	storetest.Run(t, func(_ *testing.T) store.Store {
		return memory.New()
	})
}
