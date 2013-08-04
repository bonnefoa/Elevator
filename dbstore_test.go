package elevator

import (
	"testing"
)

func TestDbstoreList(t *testing.T) {
	f := func(db_store *DbStore, db *Db) {
		lst_dbs := db_store.List()
		expected := []string{TestDb}
		if !isStringSliceEquals(lst_dbs, expected) {
			t.Fatalf("The db store should contains only [test_db]",
				lst_dbs)
		}
	}
	TemplateDbTest(t, f)
}

func TestDbstoreLoad(t *testing.T) {
	f := func(db_store *DbStore, db *Db) {
		db_store.WriteToFile()
		t.Logf("Db store is %q", db_store)
		err := db_store.ReadFromFile()
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("Db store is %q", db_store)
		db_store.Mount(db.Uid)
	}
	TemplateDbTest(t, f)
}
