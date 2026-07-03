package codec

const DefaultEncodeThreshold = 128 * 1024

var (
	Default Codec = NewZstd(ZstdConfig{
		Level: ZstdLevel(ZstdSpeedDefault),
	})
)
