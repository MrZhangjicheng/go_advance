package kms

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"
)

const (
	rootHead = "kingsoft-header"
	keyHead  = "kingsoft-key"
)

type KmsLocal struct {
	configPath   string
	secretPath   string
	rootKey      []byte
	cache        map[string][]byte // 业务密钥缓存
	mutex        sync.RWMutex
	defEncryptor SymmetricEncrypt // 默认加密算法
	geminiKey    []byte           // 更换根密钥时的新根密钥
}

func NewLocalClient(configPath, secretPath string) (*KmsLocal, error) {
	client := KmsLocal{
		configPath:   configPath,
		secretPath:   secretPath,
		rootKey:      nil,
		cache:        make(map[string][]byte, 16),
		mutex:        sync.RWMutex{},
		defEncryptor: AES256GCM{RandomNonce: true},
	}

	err := client.init()
	if err != nil {
		return nil, err
	}

	rand.Seed(time.Now().Unix())
	return &client, nil
}

func NewLocalClientWithDefEncryptor(configPath, secretPath string, defEncryptor SymmetricEncrypt) (*KmsLocal, error) {
	kmsCli, err := NewLocalClient(configPath, secretPath)
	if err != nil {
		return nil, err
	}
	kmsCli.defEncryptor = defEncryptor
	return kmsCli, nil
}

func (cli *KmsLocal) CreateKey(desc string) (string, error) {
	randID := rand.Int63n(math.MaxInt64)
	tail := fmt.Sprintf("%v", randID)

	key := cli.makeKey(keyHead, desc, tail)
	keyID, err := cli.getBusinessKeyID([]byte(key))
	if err != nil {
		return "", err
	} else {
		return keyID, nil
	}
}

func (cli *KmsLocal) CreateKeyWithMaterial(descr, material string) (string, error) {
	randID := rand.Int63n(math.MaxInt64)
	tail := fmt.Sprintf("%v", randID)

	key := cli.makeKey(keyHead, descr+material, tail)
	keyID, err := cli.getBusinessKeyID([]byte(key))
	if err != nil {
		return "", err
	} else {
		return keyID, nil
	}
}

func (cli *KmsLocal) CreateKeyWithSeed(idSeed, descr string) (string, bool, error) {
	return "", false, ErrUnsupportAPI
}

func (cli *KmsLocal) GetKeyById(id string) (*Key, error) {
	key, err := cli.getBusinessKey(id)
	if err != nil {
		return nil, err
	}

	return &Key{
		KeyId: id,
		Key:   key,
		Descr: "",
	}, nil
}

func (cli *KmsLocal) BatchGetKey(ids []string) (map[string]Key, error) {
	return nil, ErrUnsupportAPI
}

func (cli *KmsLocal) ListKeyIds() ([]string, error) {
	return nil, ErrUnsupportAPI
}

func (cli *KmsLocal) DeleteKeyById(id string) error {
	return ErrUnsupportAPI
}

func (cli *KmsLocal) Encrypt(keyId string, plainText []byte, encryptFunc func([]byte, []byte) ([]byte, error)) ([]byte, error) {
	key, err := cli.getBusinessKey(keyId)
	if err != nil {
		return nil, err
	}

	return encryptFunc(plainText, key)
}

func (cli *KmsLocal) Decrypt(keyId string, cipherText []byte, decryptFunc func([]byte, []byte) ([]byte, error)) ([]byte, error) {
	key, err := cli.getBusinessKey(keyId)
	if err != nil {
		return nil, err
	}

	return decryptFunc(cipherText, key)
}

func (cli *KmsLocal) DoEncrypt(keyId string, plainText []byte) ([]byte, error) {
	key, err := cli.getBusinessKey(keyId)
	if err != nil {
		return nil, err
	}

	return cli.defEncryptor.Encrypt(plainText, key)
}

func (cli *KmsLocal) DoDecrypt(keyId string, cipherText []byte) ([]byte, error) {
	key, err := cli.getBusinessKey(keyId)
	if err != nil {
		return nil, err
	}

	return cli.defEncryptor.Decrypt(cipherText, key)
}

func (cli *KmsLocal) readFile(filePathName string) (content string, err error) {
	b, err := ioutil.ReadFile(filePathName)
	if err != nil {
		return "", err
	}

	content = string(b)
	content = strings.Trim(content, " \r\n\t")
	return
}

func (cli *KmsLocal) makeKey(head, mid, tail string) string {
	srcData := head + mid + tail
	hash := md5.Sum([]byte(srcData))
	return fmt.Sprintf("%x", hash)
}

func (cli *KmsLocal) init() error {

	key, err := cli.makeRootKey(cli.configPath, cli.secretPath)
	if err != nil {
		return err
	}

	cli.rootKey = []byte(key)
	return nil
}

func (cli *KmsLocal) getBusinessKeyID(key []byte) (string, error) {
	id, err := cli.defEncryptor.Encrypt(key, cli.rootKey)
	if err != nil {
		return "", err
	}
	keyID := base64.StdEncoding.EncodeToString(id)

	cli.setKeyCache(keyID, key)
	return keyID, nil
}

func (cli *KmsLocal) getBusinessKey(keyID string) ([]byte, error) {
	if keyID == "" {
		return nil, ErrInvalidKeyID
	}

	// get from cache
	key, ok := cli.getKeyFromCache(keyID)
	if ok {
		return key, nil
	}

	id, err := base64.StdEncoding.DecodeString(keyID)
	if err != nil {
		return nil, err
	}

	key, err = cli.defEncryptor.Decrypt(id, cli.rootKey)
	if err != nil {
		return nil, err
	}

	// write cache
	cli.setKeyCache(keyID, key)
	return key, nil
}

func (cli *KmsLocal) getKeyFromCache(keyID string) ([]byte, bool) {
	cli.mutex.RLock()
	defer cli.mutex.RUnlock()
	key, ok := cli.cache[keyID]
	return key, ok
}

func (cli *KmsLocal) setKeyCache(keyID string, key []byte) {
	cli.mutex.Lock()
	defer cli.mutex.Unlock()
	cli.cache[keyID] = key
}

func (cli *KmsLocal) makeRootKey(configPath, secretPath string) (string, error) {
	var err error
	// read config
	configData, err := cli.readFile(configPath)
	if err != nil {
		return "", err
	}

	// read secret
	secretData, err := cli.readFile(secretPath)
	if err != nil {
		return "", err
	}

	// make root key
	key := cli.makeKey(rootHead, configData, secretData)
	return key, nil
}

func (cli *KmsLocal) Gemini(configPath, secretPath string) error {
	key, err := cli.makeRootKey(configPath, secretPath)
	if err != nil {
		return err
	}

	cli.geminiKey = []byte(key)
	return nil
}

func (cli *KmsLocal) ReEncryptWitchGemini(keyID string) (string, error) {
	if len(cli.geminiKey) == 0 {
		panic("ReEncryptWitchGemini gemini is empty!")
	}

	key, err := cli.getBusinessKey(keyID)
	if err != nil {
		return "", err
	}

	idNew, err := cli.defEncryptor.Encrypt(key, cli.geminiKey)
	if err != nil {
		return "", err
	}
	keyIDNew := base64.StdEncoding.EncodeToString(idNew)
	return keyIDNew, nil
}
