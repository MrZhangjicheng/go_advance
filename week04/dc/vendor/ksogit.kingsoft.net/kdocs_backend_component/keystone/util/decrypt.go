package util

import (
	"encoding/base64"
	"os"

	"ksogit.kingsoft.net/kgo/kms"
	"ksogit.kingsoft.net/kgo/log"
)

//解密配置数据
func Decrypt(encryptData, key string) string {
	//开启加密
	// log.Info("KAE_KMS_ENCRYPT is %s", os.Getenv("KAE_KMS_ENCRYPT"))
	if os.Getenv("KAE_KMS_ENCRYPT") == "true" {
		log.Debug("enter KAE_KMS_ENCRYPT, key is %s, encryptData is %s", key, encryptData)
		cli, errCli := kms.NewLocalClient(os.Getenv("KAE_KMS_CONFIG_PATH"), os.Getenv("KAE_KMS_SECRET_PATH"))
		if errCli != nil {
			log.Error("Init KMS Error, error is %v", errCli)
			return encryptData
		}

		// 直接用keyId对应的密钥解密数据
		plaintext, err := cli.DoDecrypt(os.Getenv("KAE_KMS_KEY_ID"), []byte(base64Decode(encryptData, key)))
		log.Error("key is %s, origin data is %s, origin data base64 is %s", key, encryptData, base64Decode(encryptData, key))
		if err != nil {
			log.Error("decrypt data failed, key is %s, origin data is %s, origin data base64 is %s,err is %v", key, encryptData, base64Decode(encryptData, key), err)
			return encryptData
		}
		return string(plaintext)
	}
	return encryptData
}

func base64Decode(src, key string) string {
	decoded, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		log.Error("base64decode failed, key is %s, val is %s, err is %v", key, src, err)
		return src
	}
	return string(decoded)
}
