package materials

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"os/user"
	"path/filepath"
	"testing"
)

func TestWalkProject(t *testing.T) {
	if true {
		return
	}
	u, _ := user.Current()
	path := filepath.Join(u.HomeDir, "Dropbox/transfers/materialscommons/WE43 Heat Treatments")
	p := Project{
		Name:   "WE43 Heat Treatments",
		Path:   path,
		Status: "Unknown",
	}

	p.Walk(nil)
}

func TestCreatedDb(t *testing.T) {
	if true {
		return
	}
	db, _ := leveldb.OpenFile("/home/gtarcea/.materials/projectdb/Synthetic Tooth.db", nil)
	defer db.Close()

	db.CompactRange(util.Range{Start: nil, Limit: nil})
	if true {
		return
	}

	iter := db.NewIterator(nil, nil)
	defer iter.Release()

	for iter.Next() {
		//key := iter.Key()
		value := iter.Value()
		var p ProjectFileInfo
		json.Unmarshal(value, &p)
		//fmt.Println(key)
		fmt.Printf("%#v\n", p)
		fmt.Println("================")
	}
}

func TestProjectFileLocationString(t *testing.T) {

}

func TestProjectFileStatusString(t *testing.T) {

}
