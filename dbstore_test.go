package elevator

import (
	"testing"
)

func TestDbStoreList(t *testing.T) {
	f := func(db_store *DbStore, db *Db) {
		lst_dbs := db_store.List()
		expected := []string{TestDb}
		if !isStringSliceEquals(lst_dbs, expected) {
			t.Fatalf("The db store should contains only [test_db]",
				lst_dbs)
		}
	}
	TemplateDbTest(t.Fatalf, f)
}
