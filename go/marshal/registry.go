package marshal

import "reflect"

var encoderRegistry map[reflect.Type]encoderFunc
var decoderRegistry map[reflect.Type]decoderFunc

func init() {
	encoderRegistry = make(map[reflect.Type]encoderFunc)
	decoderRegistry = make(map[reflect.Type]decoderFunc)
}

func RegisterEncoder(typeToEncode reflect.Type, encoder encoderFunc) {
	encoderRegistry[typeToEncode] = encoder
}

func RegisterDecoder(typeToDecode reflect.Type, decoder decoderFunc) {
	decoderRegistry[typeToDecode] = decoder
}

func GetEncoder(typeToEncode reflect.Type) encoderFunc {
	encoder,ok := encoderRegistry[typeToEncode]
	if ok {
		return encoder
	}
	return nil
}

func GetDecoder(typeToDecode reflect.Type) decoderFunc {
	decoder,ok := decoderRegistry[typeToDecode]
	if ok {
		return decoder
	}
	return nil
}