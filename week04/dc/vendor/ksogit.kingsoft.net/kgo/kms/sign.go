package kms

import (
	"bytes"
	"crypto"
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func signToken(appId, appKey string, request http.Request, bodyData []byte) (string, error) {
	buf := bytes.NewBufferString(strings.ToLower(request.Method))
	u := request.URL
	path := u.Path
	if u.RawQuery != "" {
		path += fmt.Sprintf("?%s", u.RawQuery)
	}
	buf.Write([]byte(path))
	if len(bodyData) > 0 {
		buf.Write(bodyData)
	}
	t := strconv.Itoa(int(time.Now().Unix()))
	buf.Write([]byte(t))

	h := md5.New()
	h.Write(buf.Bytes())
	header := map[string]interface{}{
		"typ":    "JWT",
		"alg":    "HS256",
		"app_id": appId,
		"s":      fmt.Sprintf("%x", h.Sum(nil)),
		"t":      t,
	}
	claims := map[string]interface{}{}
	data1, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	data2, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encode := func(data []byte) string {
		return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
	}
	sstr := strings.Join([]string{encode(data1), encode(data2)}, ".")
	hasher := hmac.New(crypto.SHA256.New, []byte(appKey))
	hasher.Write([]byte(sstr))
	sig := encode(hasher.Sum(nil))
	return strings.Join([]string{sstr, sig}, "."), nil
}
