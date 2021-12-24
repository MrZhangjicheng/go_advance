package kms

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Key struct {
	KeyId string
	Key   []byte
	Descr string
}

type KmsClient struct {
	httpcli       HttpClient
	appId, appKey string
	aesCipher     AES256GCM
	defEncryptor  SymmetricEncrypt
	scheme, host  string
	cache         map[string]Key
	mu            sync.Mutex
}

// 创建KMS-Client
func NewKmsClient(appId, appKey, kmsHost string, cli HttpClient) *KmsClient {
	var httpcli HttpClient
	httpcli = cli
	if httpcli == nil {
		httpcli = defHttpClient{cli: http.Client{}}
	}
	scheme := "https"
	if kmsHost == "kms.internal.wps.cn" {
		scheme = "http" // 内网三级域名暂时不支持https
	}
	aesGCM := AES256GCM{RandomNonce: true}
	return &KmsClient{
		httpcli:      httpcli,
		appId:        appId,
		appKey:       appKey,
		aesCipher:    aesGCM,
		defEncryptor: aesGCM,
		scheme:       scheme,
		host:         kmsHost,
		cache:        make(map[string]Key, 16),
		mu:           sync.Mutex{},
	}
}

// 创建KMS-Client
func NewKmsClientWithDefEncryptor(appId, appKey, kmsHost string, cli HttpClient, defEncryptor SymmetricEncrypt) *KmsClient {
	kmsCli := NewKmsClient(appId, appKey, kmsHost, cli)
	kmsCli.defEncryptor = defEncryptor
	return kmsCli
}

// 创建密钥
func (cli *KmsClient) CreateKey(descr string) (string, error) {
	param, _ := json.Marshal(map[string]string{
		"descr": descr,
	})
	req, err := http.NewRequest("POST", fmt.Sprintf("%s://%s/api/workkey", cli.scheme, cli.host), bytes.NewBuffer(param))
	if err != nil {
		return "", err
	}
	err = cli.sign(req, param)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-type", "application/json")
	resp, err := cli.httpcli.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", body2WebError(bodyData)
	}
	type result struct {
		KeyId  string `json:"key_id"`
		Result string `json:"result"`
	}
	var ret result
	err = json.Unmarshal(bodyData, &ret)
	if err != nil {
		return "", errors.New(string(bodyData))
	}
	return ret.KeyId, nil
}

// 创建密钥, 根据idSeed生成能和idSeed有一对一关系的keyId, 并返回
// 第一个返回参数为密钥的keyId
// 第二个返回参数为keyId是否已经存在, false说明该keyId为首次创建, true则反之
func (cli *KmsClient) CreateKeyWithSeed(idSeed, descr string) (string, bool, error) {
	param, _ := json.Marshal(map[string]string{
		"id_seed": idSeed,
		"descr":   descr,
	})
	req, err := http.NewRequest("POST", fmt.Sprintf("%s://%s/api/workkey/withseed", cli.scheme, cli.host), bytes.NewBuffer(param))
	if err != nil {
		return "", false, err
	}
	err = cli.sign(req, param)
	if err != nil {
		return "", false, err
	}
	req.Header.Set("Content-type", "application/json")
	resp, err := cli.httpcli.Do(req)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()
	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", false, err
	}
	if resp.StatusCode != 200 {
		return "", false, body2WebError(bodyData)
	}
	type result struct {
		KeyId   string `json:"key_id"`
		Result  string `json:"result"`
		Existed int64  `json:"existed"`
	}
	var ret result
	err = json.Unmarshal(bodyData, &ret)
	if err != nil {
		return "", false, errors.New(string(bodyData))
	}
	keyIdExisted := false
	if ret.Existed == 1 {
		keyIdExisted = true
	}
	return ret.KeyId, keyIdExisted, nil
}

// 创建密钥(提供密钥材料)
func (cli *KmsClient) CreateKeyWithMaterial(descr, material string) (string, error) {
	param, _ := json.Marshal(map[string]string{
		"descr":    descr,
		"material": material,
	})
	req, err := http.NewRequest("POST", fmt.Sprintf("%s://%s/api/workkey", cli.scheme, cli.host), bytes.NewBuffer(param))
	if err != nil {
		return "", err
	}
	err = cli.sign(req, param)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-type", "application/json")
	resp, err := cli.httpcli.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", body2WebError(bodyData)
	}
	type result struct {
		KeyId  string `json:"key_id"`
		Result string `json:"result"`
	}
	var ret result
	err = json.Unmarshal(bodyData, &ret)
	if err != nil {
		return "", errors.New(string(bodyData))
	}
	return ret.KeyId, nil
}

// 获取密钥
func (cli *KmsClient) GetKeyById(id string) (*Key, error) {
	key, err := cli.getKeyById(id)
	if err != nil {
		key, err = cli.getKeyById(id)
		if err != nil {
			time.Sleep(time.Millisecond * 500)
			key, err = cli.getKeyById(id)
			if err != nil {
				time.Sleep(time.Second)
				key, err = cli.getKeyById(id)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return key, nil
}

func (cli *KmsClient) getKeyById(id string) (*Key, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s://%s/api/workkey?id=%s", cli.scheme, cli.host, id), nil)
	if err != nil {
		return nil, err
	}
	err = cli.sign(req, nil)
	if err != nil {
		return nil, err
	}
	resp, err := cli.httpcli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, body2WebError(bodyData)
	}
	type result struct {
		KeyData  string `json:"key_data"`
		KeyId    string `json:"key_id"`
		KeyDescr string `json:"key_descr"`
		Result   string `json:"result"`
	}
	var ret result
	err = json.Unmarshal(bodyData, &ret)
	if err != nil {
		return nil, errors.New(string(bodyData))
	}
	keyData, err := cli.aesCipher.DecodeDecrypt(ret.KeyData, cli.appKey)
	if err != nil {
		return nil, err
	}
	plainData, err := base64.RawStdEncoding.DecodeString(keyData)
	if err != nil {
		return nil, err
	}
	return &Key{
		KeyId: ret.KeyId,
		Key:   plainData,
		Descr: ret.KeyDescr,
	}, nil
}

// 批量获取密钥
func (cli *KmsClient) BatchGetKey(ids []string) (map[string]Key, error) {
	body := "ids=" + strings.Join(ids, ",")
	req, err := http.NewRequest("POST", fmt.Sprintf("%s://%s/api/workkey/batch", cli.scheme, cli.host), strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	err = cli.sign(req, []byte(body))
	if err != nil {
		return nil, err
	}
	resp, err := cli.httpcli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, body2WebError(bodyData)
	}

	type keyResult struct {
		KeyData  string `json:"key_data"`
		KeyId    string `json:"key_id"`
		KeyDescr string `json:"key_descr"`
	}
	type result struct {
		Keys   map[string]keyResult `json:"keys"`
		Result string               `json:"result"`
	}
	var ret result
	err = json.Unmarshal(bodyData, &ret)
	if err != nil {
		return nil, errors.New(string(bodyData))
	}
	kMap := make(map[string]Key, len(ids))
	for _, id := range ids {
		k, ok := ret.Keys[id]
		if ok {
			keyData, err := cli.aesCipher.DecodeDecrypt(k.KeyData, cli.appKey)
			if err != nil {
				return nil, err
			}
			plainData, err := base64.RawStdEncoding.DecodeString(keyData)
			kMap[id] = Key{
				KeyId: id,
				Key:   plainData,
				Descr: k.KeyDescr,
			}
		}
	}
	return kMap, nil
}

// 获取密钥ID列表
func (cli *KmsClient) ListKeyIds() ([]string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s://%s/api/workkey/list", cli.scheme, cli.host), nil)
	if err != nil {
		return nil, err
	}
	err = cli.sign(req, nil)
	if err != nil {
		return nil, err
	}
	resp, err := cli.httpcli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, body2WebError(bodyData)
	}

	type result struct {
		KeyIds []string `json:"key_ids"`
		Result string   `json:"result"`
	}
	var ret result
	err = json.Unmarshal(bodyData, &ret)
	if err != nil {
		return nil, errors.New(string(bodyData))
	}
	return ret.KeyIds, nil
}

// 删除密钥
func (cli *KmsClient) DeleteKeyById(id string) error {
	param, _ := json.Marshal(map[string]string{
		"id": id,
	})
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s://%s/api/workkey", cli.scheme, cli.host), bytes.NewBuffer(param))
	if err != nil {
		return err
	}
	err = cli.sign(req, param)
	if err != nil {
		return err
	}
	req.Header.Set("Content-type", "application/json")
	resp, err := cli.httpcli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return body2WebError(bodyData)
	}
	return nil
}

func (cli *KmsClient) SetHttpClient(client HttpClient) {
	cli.httpcli = client
}

// 用keyId对应密钥加密数据
func (cli *KmsClient) Encrypt(keyId string, plaintext []byte, encryptFunc func([]byte, []byte) ([]byte, error)) ([]byte, error) {
	cli.mu.Lock()
	defer cli.mu.Unlock()
	keyCache, ok := cli.cache[keyId]
	if ok {
		return encryptFunc(plaintext, keyCache.Key)
	}
	key, err := cli.GetKeyById(keyId)
	if err != nil {
		return nil, err
	}
	cli.cache[keyId] = *key
	return encryptFunc(plaintext, key.Key)
}

// 用keyId对应密钥对数据进行解密
func (cli *KmsClient) Decrypt(keyId string, ciphertext []byte, decryptFunc func([]byte, []byte) ([]byte, error)) ([]byte, error) {
	cli.mu.Lock()
	defer cli.mu.Unlock()
	keyCache, ok := cli.cache[keyId]
	if ok {
		return decryptFunc(ciphertext, keyCache.Key)
	}
	key, err := cli.GetKeyById(keyId)
	if err != nil {
		return nil, err
	}
	cli.cache[keyId] = *key
	return decryptFunc(ciphertext, key.Key)
}

// sdk提供的默认的加密算法
func (cli *KmsClient) DoEncrypt(keyId string, plaintext []byte) ([]byte, error) {
	return cli.Encrypt(keyId, plaintext, cli.defEncryptor.Encrypt)
}

// sdk提供的默认的解密算法
func (cli *KmsClient) DoDecrypt(keyId string, ciphertext []byte) ([]byte, error) {
	return cli.Decrypt(keyId, ciphertext, cli.defEncryptor.Decrypt)
}

func (cli KmsClient) sign(request *http.Request, bodyData []byte) error {
	token, err := signToken(cli.appId, cli.appKey, *request, bodyData)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return nil
}

func body2WebError(bodyData []byte) WebErr {
	var werr WebErr
	err := json.Unmarshal(bodyData, &werr)
	if err != nil {
		return WebErr{
			Result: string(bodyData),
		}
	}
	return werr
}

type WebErr struct {
	Result string `json:"result"`
	Msg    string `json:"msg"`
}

func (err WebErr) Error() string {
	return fmt.Sprintf("%s:%s", err.Result, err.Msg)
}

type defHttpClient struct {
	cli http.Client
}

func (dCli defHttpClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := dCli.cli.Do(req)
	if err != nil {
		// retry
		resp, err = dCli.cli.Do(req)
		if err != nil {
			time.Sleep(time.Millisecond * 200)
			resp, err = dCli.cli.Do(req)
			if err != nil {
				time.Sleep(time.Millisecond * 400)
				resp, err = dCli.cli.Do(req)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return resp, nil
}
