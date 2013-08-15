package store

import (
	"testing"
	"reflect"
)

func TestDbstoreList(t *testing.T) {
    env := setupEnv(t)
    defer env.destroy()

    lstDbs := env.DbStore.List()
    expected := []string{TestDb}
    if !reflect.DeepEqual(lstDbs, expected) {
        t.Fatal("The db store should contains only [test_db]", lstDbs)
    }
}

func TestDbstoreLoad(t *testing.T) {
    env := setupEnv(t)
    defer env.destroy()

    env.DbStore.WriteToFile()
    t.Logf("Db store is %q", env.DbStore)
    err := env.DbStore.Load()
    if err != nil {
        t.Fatal(err)
    }
    t.Logf("Db store is %q", env.DbStore)
    env.DbStore.Mount(env.db.UID)
}
