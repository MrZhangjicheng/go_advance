/*
KMS-SDK

example:

	// 创建Kms客户端实例, HttpClient可以自己实现, 方便审计, 为nil则默认创建一个http.Client作为httpClient
    // kmsHost: 如果服务部署在KAE则推荐用"kms.internal.wps.cn", 否则"kms.wps.cn"
	cli := kms.NewKmsClient("myappid", "myappkey", "kms.internal.wps.cn", nil)

	// 直接用keyId对应的密钥加密数据
	ciphertext, err := cli.DoEncrypt(keyId, plaintext)

    // 直接用keyId对应的密钥解密数据
	plaintext, err := cli.DoDecrypt(keyId, ciphertext)

	// 获取密钥
	myKey, err := cli.GetKeyById(keyId)
	myKey.KeyId // key的id : string
	myKey.Key   // key的明文 : []byte
	myKey.Descr // key的用途 : string

	// 获取密钥(批量): 返回map[keyid]Key
	myKeys, err := cli.BatchGetKey([]string{keyId1, keyId2}) // myKey: map[string]Key

	// 获取myappid下的keyId列表
	keyIds, err := cli.ListKeyIds() // keyId列表 : []string

	// 创建密钥,返回KeyId; 其中keyId2==keyId3, 因为用了同一个seed: "phone_cipher"
    // existed1==false(原因为首次创建) , existed2==true
	keyId2, existed1, err := cli.CreateKeyWithSeed("phone_cipher", "用户手机加密")
	keyId3, existed2, err = cli.CreateKeyWithSeed("phone_cipher", "用户手机加密")
*/
package kms
