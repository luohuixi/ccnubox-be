package crawler

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
)

const (
	postgraduateURL      = "https://grd.ccnu.edu.cn"
	publicKeyURL         = postgraduateURL + "/yjsxt/xtgl/login_getPublicKey.html"
	loginPostgraduateURL = postgraduateURL + "/yjsxt/xtgl/login_slogin.html"
)

type PostGraduate struct {
}

func NewPostGraduate() *PostGraduate {
	return &PostGraduate{}
}

// FetchPublicKey 1. 获取账号密码的加密秘钥
func (c *PostGraduate) FetchPublicKey(client *http.Client) (*rsa.PublicKey, error) {
	req, err := http.NewRequest("GET", publicKeyURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", postgraduateURL+"/yjsxt/")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data rsaPublicKeyResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return parseRSAPublicKey(data.Modulus, data.Exponent)
}

// LoginPostgraduateSystem 2.登陆研究生院
func (c *PostGraduate) LoginPostgraduateSystem(client *http.Client, username, password string, pubKey *rsa.PublicKey) (*http.Client, error) {

	encPwd, err := encryptPasswordJSStyle(password, pubKey)
	if err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Set("csrftoken", "")
	form.Set("yhm", username)
	form.Set("mm", encPwd)
	form.Set("hidMm", encPwd)

	req, err := http.NewRequest("POST", loginPostgraduateURL, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", postgraduateURL+"/yjsxt/")
	req.Header.Set("Origin", postgraduateURL)
	req.Header.Set("Host", "grd.ccnu.edu.cn")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// TODO 效率有点低,暂时先这样
	if strings.Contains(string(body), "用户名或密码不正确") {
		return nil, INCorrectPASSWORD
	}

	return client, nil
}

type rsaPublicKeyResponse struct {
	Modulus  string `json:"modulus"`
	Exponent string `json:"exponent"`
}

func parseRSAPublicKey(modBase64, expBase64 string) (*rsa.PublicKey, error) {
	modBytes, err := base64.StdEncoding.DecodeString(modBase64)
	if err != nil {
		return nil, fmt.Errorf("modulus decode error: %v", err)
	}
	expBytes, err := base64.StdEncoding.DecodeString(expBase64)
	if err != nil {
		return nil, fmt.Errorf("exponent decode error: %v", err)
	}
	modulus := new(big.Int).SetBytes(modBytes)
	exponent := new(big.Int).SetBytes(expBytes)

	return &rsa.PublicKey{
		N: modulus,
		E: int(exponent.Int64()),
	}, nil
}

func encryptPasswordJSStyle(password string, pubKey *rsa.PublicKey) (string, error) {
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, []byte(password))
	if err != nil {
		return "", err
	}
	hexStr := hex.EncodeToString(encrypted)
	hexBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(hexBytes), nil
}
