package gsframe

import (
	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/bstree"
)

func NewBSTree() gsinf.IBSTree {
	return bstree.NewBSTree()
}
