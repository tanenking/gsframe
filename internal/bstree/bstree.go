package bstree

import (
	"strings"
	"sync"

	"github.com/tanenking/gsframe/gsinf"
)

type bstNodeBase struct {
	Value  gsinf.IBSTData
	dwSize uint32
}

type bstNode struct {
	bstNodeBase
	ptLeft  *bstNode
	ptRight *bstNode
}

func newBSTNode() *bstNode {
	return &bstNode{ptLeft: nil, ptRight: nil}
}

type BSTree struct {
	sync.Mutex
	ptRoot  *bstNode
	mapData map[string]*bstNode
}

func NewBSTree() *BSTree {
	bst := &BSTree{
		Mutex:   sync.Mutex{},
		ptRoot:  nil,
		mapData: map[string]*bstNode{},
	}
	return bst
}

// 设置新值, 覆盖旧值, 返回名次, 从1开始
func (bst *BSTree) Set(value gsinf.IBSTData) int32 {
	bst.Lock()
	defer bst.Unlock()

	key := value.Key()
	if tmp, ok := bst.mapData[key]; ok {
		bst.delete(key, tmp.Value)
	}

	nd := bst.insert(value)
	bst.mapData[key] = nd
	return int32(bst.lessCount(value)) + 1
}

func (bst *BSTree) Delete(key string) {
	bst.Lock()
	defer bst.Unlock()

	if tmp, ok := bst.mapData[key]; ok {
		bst.delete(key, tmp.Value)
		delete(bst.mapData, key)
	}
}

// 返回排行名次和数据, 不存在返回 0
func (bst *BSTree) GetRankData(key string) (int32, gsinf.IBSTData) {
	bst.Lock()
	defer bst.Unlock()

	if tmp, ok := bst.mapData[key]; ok {
		return int32(bst.lessCount(tmp.Value)) + 1, tmp.Value
	}
	return 0, nil
}
func (bst *BSTree) GetDataFromRank(rank int32) gsinf.IBSTData { //通过名次返回数据
	if rank <= 0 || rank > int32(bst.Size()) {
		return nil
	}
	bst.Lock()
	defer bst.Unlock()
	var rankData gsinf.IBSTData
	bst.EnumRankList(func(nRank int32, data gsinf.IBSTData) bool {
		if nRank == rank {
			rankData = data
			return false
		}
		return true
	})
	return rankData
}

// 获取所有数据, 已排序, 0代表取所有数据
func (bst *BSTree) GetRankList(count uint32) []gsinf.IBSTData {
	bst.Lock()
	defer bst.Unlock()

	return bst.getRankList(count)
}
func (bst *BSTree) EnumRankList(fn func(nRank int32, val gsinf.IBSTData) bool) { //枚举排行榜 return false 中断枚举
	bst.Lock()
	defer bst.Unlock()
	if bst.ptRoot == nil {
		return
	}
	var fnEnum func(ptRoot *bstNode)
	var nRank int32 = 0
	var finished bool = false
	fnEnum = func(ptRoot *bstNode) {
		if ptRoot == nil {
			return
		}
		if finished {
			return
		}
		fnEnum(ptRoot.ptLeft)
		nRank++
		if !fn(nRank, ptRoot.Value) {
			finished = true
			return
		}
		fnEnum(ptRoot.ptRight)
	}
	fnEnum(bst.ptRoot)
}
func (bst *BSTree) Size() uint32 {
	if bst.ptRoot == nil {
		return 0
	}

	return bst.ptRoot.dwSize
}

func (bst *BSTree) ClearBSTree() {
	bst.Lock()
	defer bst.Unlock()

	clearNode(bst.ptRoot)
	bst.ptRoot = nil
	bst.mapData = map[string]*bstNode{}
}

///////////////////////////////////////////////////////////////////////////////

// 获取所有数据, 已排序, 0代表取所有数据
func (bst *BSTree) getRankList(maxcount uint32) []gsinf.IBSTData {
	if maxcount <= 0 {
		maxcount = bst.Size()
	}
	var count uint32 = 0
	data := make([]gsinf.IBSTData, 0, maxcount)
	getAllData(bst.ptRoot, &data, &count, &maxcount)
	return data
}

// 树中键值小于value的结点个数
func (bst *BSTree) lessCount(value gsinf.IBSTData) uint32 {
	return sbtLessCount(bst, bst.ptRoot, value)
}

func (bst *BSTree) insert(value gsinf.IBSTData) *bstNode {
	pNode := newBSTNode()
	pNode.dwSize = 0
	pNode.Value = value
	sbtInsert(bst, &bst.ptRoot, pNode)
	return pNode
}

func (bst *BSTree) delete(key string, value gsinf.IBSTData) bool {
	if pNode := sbtDelete(bst, &bst.ptRoot, key, value, false); pNode != nil {
		return true
	}
	return false
}

func getAllData(ptRoot *bstNode, data *[]gsinf.IBSTData, count, maxcount *uint32) {
	if ptRoot == nil {
		return
	}

	getAllData(ptRoot.ptLeft, data, count, maxcount)
	if *count >= *maxcount {
		return
	}
	*count++
	*data = append(*data, ptRoot.Value)
	getAllData(ptRoot.ptRight, data, count, maxcount)
}

func sbtLessCount(bst *BSTree, pNode *bstNode, value gsinf.IBSTData) uint32 {
	if pNode == nil {
		return 0
	}

	if pNode.Value.Compare(value) {
		//当前节点小于指定值,则当前节点的子左节点树,都小于指定值,子右节点树不确定
		return getSize(pNode.ptLeft) + 1 + sbtLessCount(bst, pNode.ptRight, value)
	}
	return sbtLessCount(bst, pNode.ptLeft, value)
}

func clearNode(pNode *bstNode) {
	if pNode == nil {
		return
	}

	clearNode(pNode.ptLeft)
	clearNode(pNode.ptRight)
}

func getSize(pNode *bstNode) uint32 {
	if pNode != nil {
		return pNode.dwSize
	}

	return 0
}

func sbtInsert(bst *BSTree, pTree **bstNode, pNode *bstNode) {
	if *pTree == nil {
		pNode.dwSize = 1
		*pTree = pNode
		return
	}

	(*pTree).dwSize++
	bIsLeft := pNode.Value.Compare((*pTree).Value)
	if bIsLeft {
		sbtInsert(bst, &(*pTree).ptLeft, pNode)
	} else {
		sbtInsert(bst, &(*pTree).ptRight, pNode)
	}

	maintain(pTree, !bIsLeft)
}

func sbtDelete(bst *BSTree, pNode **bstNode, key string, value gsinf.IBSTData, bIsFind bool) (record *bstNode) {
	// 查找到key所在的节点，然后用该节点左子树中最大值节点来替换掉需要删除的节点
	if *pNode == nil {
		return nil
	}

	// 只支持每次删除掉一个节点
	if bIsFind && (*pNode).ptRight == nil {
		// 查找最右的子树中最右的节点来替代删除的节点
		record = *pNode
		*pNode = (*pNode).ptLeft
		return record
	}

	if !bIsFind && strings.Compare(key, (*pNode).Value.Key()) == 0 {
		// 如果查询到是相等的节点
		if (*pNode).dwSize == 1 {
			// 叶子节点，需要删除掉此节点
			record = *pNode
			*pNode = nil
			return record
		}

		if (*pNode).dwSize == 2 {
			// 单枝节点，需要子节点继承被删除的节点
			if (*pNode).ptLeft != nil {
				record = (*pNode).ptLeft
				(*pNode).ptLeft = nil
			} else {
				record = (*pNode).ptRight
				(*pNode).ptRight = nil
			}
		} else {
			// 当前节点的左子树的最大值作为pNode的代替节点
			record = sbtDelete(bst, &(*pNode).ptLeft, key, value, true)
		}

		if record != nil {
			(*pNode).Value = record.Value
		}

		(*pNode).dwSize--
		maintain(pNode, true)
	} else if !bIsFind && ((*pNode).ptLeft != nil && strings.Compare(key, ((*pNode).ptLeft).Value.Key()) == 0) {
		record = sbtDelete(bst, &(*pNode).ptLeft, key, value, false)
		if record != nil {
			(*pNode).dwSize--
		}
	} else if !bIsFind && ((*pNode).ptRight != nil && strings.Compare(key, ((*pNode).ptRight).Value.Key()) == 0) {
		record = sbtDelete(bst, &(*pNode).ptRight, key, value, false)
		if record != nil {
			(*pNode).dwSize--
		}
	} else if !bIsFind && value.Compare((*pNode).Value) {
		record = sbtDelete(bst, &(*pNode).ptLeft, key, value, false)
		if record != nil {
			(*pNode).dwSize--
		}
	} else {
		record = sbtDelete(bst, &(*pNode).ptRight, key, value, bIsFind)
		if record != nil {
			(*pNode).dwSize--
		}
	}

	if record != nil {
		maintain(pNode, !bIsFind && value.Compare((*pNode).Value))
	}

	return record
}

func leftRotate(pNode **bstNode) {
	pRight := (*pNode).ptRight
	(*pNode).ptRight = pRight.ptLeft
	pRight.ptLeft = *pNode
	pRight.dwSize = (*pNode).dwSize
	(*pNode).dwSize = getSize((*pNode).ptLeft) + getSize((*pNode).ptRight) + 1
	*pNode = pRight
}

func rightRotate(pNode **bstNode) {
	pLeft := (*pNode).ptLeft
	(*pNode).ptLeft = pLeft.ptRight
	pLeft.ptRight = *pNode
	pLeft.dwSize = (*pNode).dwSize
	(*pNode).dwSize = getSize((*pNode).ptLeft) + getSize((*pNode).ptRight) + 1
	*pNode = pLeft
}

func maintain(pNode **bstNode, bIsRightDeeper bool) {
	if *pNode == nil {
		return
	}

	if !bIsRightDeeper {
		if (*pNode).ptLeft == nil {
			return
		}

		var dwRSize uint32 = getSize((*pNode).ptRight)
		if getSize((*pNode).ptLeft.ptLeft) > dwRSize {
			rightRotate(pNode)
		} else if getSize((*pNode).ptLeft.ptRight) > dwRSize {
			leftRotate(&(*pNode).ptLeft)
			rightRotate(pNode)
		} else {
			return
		}

		maintain(&(*pNode).ptLeft, false)
	} else {
		if (*pNode).ptRight == nil {
			return
		}

		var dwLSize uint32 = getSize((*pNode).ptLeft)
		if getSize((*pNode).ptRight.ptRight) > dwLSize {
			leftRotate(pNode)
		} else if getSize((*pNode).ptRight.ptLeft) > dwLSize {
			rightRotate(&(*pNode).ptRight)
			leftRotate(pNode)
		} else {
			return
		}

		maintain(&(*pNode).ptRight, true)
	}

	maintain(pNode, false)
	maintain(pNode, true)
}

/////////////////////////////////////////////////////////////////////////////////////////////
