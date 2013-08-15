package store

import (
	"testing"
    "reflect"
)

func TestRoutine(t *testing.T) {
    env := setupEnv(t)
    defer env.destroy()

    for i, tt := range testOperationDatas {
        request := &Request{ dbUID:env.Db.UID, Command:tt.op,
            Args:tt.request, requestType: typeDb}
        res, err := env.DbStore.HandleRequest(request)
        if !reflect.DeepEqual(err, tt.expectedError) {
            t.Fatalf("%d, expected status %v, got %v", i,
            tt.expectedError, err)
        }
        if !reflect.DeepEqual(res, tt.data) {
            t.Fatalf("expected %v, got %v", tt.data, res)
        }
    }
}
