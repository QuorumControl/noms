package marshal

import (
	"reflect"
	"github.com/attic-labs/noms/go/types"
	"time"
	"fmt"
	"strconv"
)

func init() {
	RegisterEncoder(reflect.TypeOf(time.Time{}), timeEncoder)
	RegisterDecoder(reflect.TypeOf(time.Time{}), timeDecoder)
}

func timeEncoder(v reflect.Value) types.Value {
	if v.IsNil() {
		return types.String("0")
	}

	nanosSinceEpoch := v.Interface().(*time.Time).UnixNano()

	return types.String(strconv.FormatInt(nanosSinceEpoch, 10))
}

func timeDecoder(v types.Value, rv reflect.Value) {
	if n, ok := v.(types.String); ok {

		nanosSinceEpoch, err := strconv.ParseInt(string(n), 10, 0)
		if err != nil {
			panic(&UnmarshalTypeMismatchError{v, rv.Type(), "string for nanosSinceEpoch was not in a number format"})
		}

		if reflect.ValueOf(int64(0)).OverflowInt(nanosSinceEpoch){
			panic(overflowError(types.Number(nanosSinceEpoch), rv.Type()))
		}

		fmt.Printf("rv is: %+v, kind: %v, type: %v\n", rv, rv.Kind(), rv.Type())

		oldTime := time.Unix(0, nanosSinceEpoch)

		newVal := reflect.ValueOf(&oldTime)

		if !rv.IsNil() {
			newVal = reflect.Indirect(rv)
		}


		fmt.Printf("adding duration: %d", nanosSinceEpoch)
		newVal.Interface().(*time.Time).Add(time.Duration(nanosSinceEpoch))
		rv.Set(newVal)
	} else {
		panic(&UnmarshalTypeMismatchError{v, rv.Type(), ""})
	}
}