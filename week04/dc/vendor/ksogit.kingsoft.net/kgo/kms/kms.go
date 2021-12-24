package kms

type KmsAccess interface {
	CreateKey(desc string) (string, error)
	CreateKeyWithMaterial(descr, material string) (string, error)
	CreateKeyWithSeed(idSeed, descr string) (string, bool, error)
	GetKeyById(id string) (*Key, error)
	BatchGetKey(ids []string) (map[string]Key, error)
	ListKeyIds() ([]string, error)
	DeleteKeyById(id string) error
	Encrypt(keyId string, plainText []byte, encryptFunc func([]byte, []byte) ([]byte, error)) ([]byte, error)
	Decrypt(keyId string, cipherText []byte, decryptFunc func([]byte, []byte) ([]byte, error)) ([]byte, error)
	DoEncrypt(keyId string, plainText []byte) ([]byte, error)
	DoDecrypt(keyId string, cipherText []byte) ([]byte, error)
}
