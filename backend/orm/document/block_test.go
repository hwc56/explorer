package document

import (
	"testing"
)

func TestGetBlockListByOffsetAndSize(t *testing.T) {

	blocks, err := Block{}.GetBlockListByOffsetAndSize(100, 50)

	if err != nil {
		t.Error(err)
	}

	for k, v := range blocks {
		t.Logf("k: %v  v: %v \n", k, v)
	}

}

func TestQueryBlockHeightTimeHashByHeight(t *testing.T) {

	block, err := Block{}.QueryBlockHeightTimeHashByHeight(100)
	if err != nil {
		t.Error(err)
	}

	t.Logf("height: 100 block: %v\n", block)
}


func TestGetRecentBlockList(t *testing.T) {

	blockList, err := Block{}.GetRecentBlockList()
	if err != nil {
		t.Error(err)
	}

	for k, v := range blockList {
		t.Logf("k: %v  v: %v\n", k, v)
	}

}

