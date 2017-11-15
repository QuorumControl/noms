package integration

import (
	"fmt"
	"github.com/attic-labs/noms/go/datas"
	"github.com/attic-labs/noms/go/types"
	"time"
	"math/rand"
	"github.com/attic-labs/noms/go/marshal"
	"github.com/spf13/afero"
	"testing"
	"github.com/attic-labs/noms/go/chunks"
	"os"
	"runtime/trace"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

const DefaultMemTableSize = 8 * (1 << 20) // 8MB


type IdentityLike struct {
	RootCertificate Certificate
	IntermediateCertificate Certificate
	Devices map[string]DeviceLike
	Metadata map[string]string
	UUID string
}

func NewIdentityLike() *IdentityLike {
	initialDevice := NewDeviceLike()
	return &IdentityLike{
		UUID: RandString(50),
		RootCertificate: NewCertificate(),
		IntermediateCertificate: NewCertificate(),
		Devices: map[string]DeviceLike{
			initialDevice.UUID: initialDevice,
		},
		Metadata: map[string]string{
			"EncryptedRootKey": "",
		},
	}
}

type DeviceLike struct {
	UUID string
	Certificate Certificate
	Description string
	Metadata    map[string]string
}

func NewDeviceLike() DeviceLike {
	return DeviceLike{
		UUID: RandString(100),
		Certificate: NewCertificate(),
		Description: RandString(200),
	}
}

type Certificate struct {
	Pem string
}

func NewCertificate() Certificate {
	return Certificate{Pem: RandString(2048)}
}



func getIdentities(ds datas.Dataset) types.Map {
	hv, ok := ds.MaybeHeadValue()
	if ok {
		return hv.(types.Map)
	}
	return types.NewMap(ds.Database())
}



func Save(ds datas.Dataset, id *IdentityLike) error {
	val,err := marshal.Marshal(ds.Database(), *id)
	if err != nil {
		return fmt.Errorf("error marshaling: %v", err)
	}


	_, err = ds.Database().CommitValue(ds, getIdentities(ds).Edit().Set(types.String(id.UUID), val).Map())

	if err != nil {
		return fmt.Errorf("error committing values: %v", err)
	}

	return nil
}

func TestSaveAndUpdate(t *testing.T) {
	fs := afero.NewOsFs()
	fs.RemoveAll("tmp/noms")
	fs.MkdirAll("tmp/noms", 0755)

	traceFile,_ := fs.OpenFile("tmp/trace", os.O_CREATE | os.O_WRONLY, 0755)
	defer traceFile.Close()

	alice := NewIdentityLike()
	newDevice := NewDeviceLike()

	//sp := datas.NewDatabase(nbs.NewLocalStore("tmp/noms", DefaultMemTableSize))
	sp := datas.NewDatabase((&chunks.MemoryStorage{}).NewView())
	defer sp.Close()

	err := Save(sp.GetDataset("identities"), alice)

	if err != nil {
		t.Fatalf("error saving: %v", err)
	}

	//update alice

	alice.Metadata = map[string]string{"myUpdate": "another thing"}

	alice.Devices[newDevice.UUID] = newDevice

	dataset := sp.GetDataset("identities")

	trace.Start(traceFile)
	err = Save(dataset, alice)
	trace.Stop()

	if err != nil {
		t.Fatalf("error getting fields: %v", err)
	}

	hv, ok := sp.GetDataset("identities").MaybeHeadValue()
	if ok {
		people := hv.(types.Map)

		dbAliceMarshaled := people.Get(types.String(alice.UUID))

		dbAlice := &IdentityLike{}

		err = marshal.Unmarshal(dbAliceMarshaled, dbAlice)

		if err != nil {
			t.Fatalf("Error unmarshaling: %v", err)
		}

		if alice.UUID != dbAlice.UUID {
			t.Errorf("alices were not equal\n\n alice:\n %v \n\ndbAlice:\n %v", alice, dbAlice)
		}

	} else {
		t.Fatalf("no head value")
	}

}