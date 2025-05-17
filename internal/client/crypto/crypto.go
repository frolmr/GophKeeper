package crypto

type EncryptionParams struct {
	Time       uint32 // Iterations
	Memory     uint32 // Memory in KiB
	Threads    uint8  // Number of threads
	SaltLength uint32 // Salt length in bytes
	KeyLength  uint32 // Derived key length in bytes
}

var DefaultParams = &EncryptionParams{
	Time:       3,
	Memory:     64 * 1024,
	Threads:    4,
	SaltLength: 16,
	KeyLength:  32,
}

type CryptoService struct {
}

func NewCryptoService() *CryptoService {
	return &CryptoService{}
}
