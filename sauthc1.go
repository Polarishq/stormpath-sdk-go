package stormpath

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const nl = "\n"

//Authenticate generates the proper authentication header for the SAuthc1 algorithm use by Stormpath
func Authenticate(req *http.Request, payload *bytes.Reader, date time.Time, credentials *Credentials, nonce string) {
	timestamp := date.Format("20060102T150405Z0700")
	dateStamp := date.Format("20060102")
	req.Header.Set("Host", req.URL.Host)
	req.Header.Set("X-Stormpath-Date", timestamp)

	signedHeadersString := signedHeadersString(req.Header)

	canonicalRequest :=
		req.Method +
			nl +
			canonicalizeResourcePath(req.URL.Path) +
			nl +
			canonicalizeQueryString(req) +
			nl +
			canonicalizeHeadersString(req.Header) +
			nl +
			signedHeadersString +
			nl +
			hex.EncodeToString(hash(payload))

	id := credentials.Id + "/" + dateStamp + "/" + nonce + "/" + "sauthc1_request"

	canonicalRequestHashHex := hex.EncodeToString(hash(bytes.NewReader([]byte(canonicalRequest))))

	stringToSign :=
		"HMAC-SHA-256" +
			nl +
			timestamp +
			nl +
			id +
			nl +
			canonicalRequestHashHex

	secret := []byte("SAuthc1" + credentials.Secret)
	singDate := sing(dateStamp, secret)
	singNonce := sing(nonce, singDate)
	signing := sing("sauthc1_request", singNonce)

	signature := sing(stringToSign, signing)
	signatureHex := hex.EncodeToString(signature)

	authorizationHeader :=
		createNameValuePair("sauthc1Id", id) + ", " +
			createNameValuePair("sauthc1SignedHeaders", signedHeadersString) + ", " +
			createNameValuePair("sauthc1Signature", signatureHex)

	req.Header.Set("Authorization", authorizationHeader)
}

func createNameValuePair(name string, value string) string {
	return name + "=" + value
}

func encodeURL(value string, path bool, canonical bool) string {
	if value == "" {
		return ""
	}

	encoded := url.QueryEscape(value)

	if canonical {
		encoded = strings.Replace(encoded, "+", "%20", -1)
		encoded = strings.Replace(encoded, "*", "%2A", -1)
		encoded = strings.Replace(encoded, "%7E", "~", -1)

		if path {
			encoded = strings.Replace(encoded, "%2F", "/", -1)
		}
	}

	return encoded
}

func canonicalizeQueryString(req *http.Request) string {
	stringBuffer := bytes.NewBufferString("")
	queryValues := req.URL.Query()

	keys := sortedMapKeys(queryValues)

	for _, k := range keys {
		key := encodeURL(k, false, true)
		v := queryValues[k]
		for _, vv := range v {
			value := encodeURL(vv, false, true)

			if stringBuffer.Len() > 0 {
				stringBuffer.WriteString("&")
			}

			stringBuffer.WriteString(key + "=" + value)
		}
	}

	return stringBuffer.String()
}

func canonicalizeResourcePath(path string) string {
	if len(path) == 0 {
		return "/"
	} else {
		return encodeURL(path, true, true)
	}
}

func canonicalizeHeadersString(headers http.Header) string {
	stringBuffer := bytes.NewBufferString("")

	keys := sortedMapKeys(headers)

	for _, k := range keys {
		stringBuffer.WriteString(strings.ToLower(k))
		stringBuffer.WriteString(":")

		first := true

		for _, v := range headers[k] {
			if !first {
				stringBuffer.WriteString(",")
			}
			stringBuffer.WriteString(v)
			first = false
		}
		stringBuffer.WriteString(nl)
	}

	return stringBuffer.String()
}

func signedHeadersString(headers http.Header) string {
	stringBuffer := bytes.NewBufferString("")

	keys := sortedMapKeys(headers)

	for _, k := range keys {
		if stringBuffer.Len() > 0 {
			stringBuffer.WriteString(";")
		}
		stringBuffer.WriteString(strings.ToLower(k))
	}

	return stringBuffer.String()
}

func sortedMapKeys(m map[string][]string) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func hash(data *bytes.Reader) []byte {
	hash := sha256.New()
	io.Copy(hash, data)
	data.Seek(0, 0)
	return hash.Sum(nil)
}

func sing(data string, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}
