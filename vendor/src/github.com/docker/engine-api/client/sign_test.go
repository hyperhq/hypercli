package client

import (
	"net/http"
	"net/url"
	"testing"
)

const (
	AccessKey string = "0PN5J17HBGZHT7JJ3X82"
	SecretKey string = "uV3F3YluFJax1cknvbcGwgjvx4QpvB+leU8dUj2o"
)

func TestSignExampleObject(t *testing.T) {
	var testCases = []struct {
		Method    string
		Path      string
		Headers   map[string]string
		Params    map[string]string
		Signature string
	}{
		// User the test cases of AWS S3 to verify
		{
			"GET",
			"/johnsmith/photos/puppy.jpg",
			map[string]string{
				"Host": "johnsmith.s3.amazonaws.com",
				"Date": "Tue, 27 Mar 2007 19:36:42 +0000"},
			map[string]string{},
			"xXjDGYUmKxnwqr5KXNPGldn5LbA=",
		},
		{
			"PUT",
			"/johnsmith/photos/puppy.jpg",
			map[string]string{
				"Host":           "johnsmith.s3.amazonaws.com",
				"Date":           "Tue, 27 Mar 2007 21:15:45 +0000",
				"Content-Type":   "image/jpeg",
				"Content-Length": "94328"},
			map[string]string{},
			"hcicpDDvL9SsO6AkvxqmIWkmOuQ=",
		},
		{
			"GET",
			"/johnsmith/",
			map[string]string{
				"Host":       "johnsmith.s3.amazonaws.com",
				"Date":       "Tue, 27 Mar 2007 19:42:41 +0000",
				"User-Agent": "Mozilla/5.0"},
			map[string]string{
				"prefix":   "photos",
				"max-keys": "50",
				"marker":   "puppy"},
			"jsRt/rhG+Vtp88HrYL706QhE4w4=",
		},
		{
			"GET",
			"/johnsmith/",
			map[string]string{
				"Host": "johnsmith.s3.amazonaws.com",
				"Date": "Tue, 27 Mar 2007 19:44:46 +0000"},
			map[string]string{
				"acl": ""},
			"thdUi9VAkzhkniLj96JIrOPGi0g=",
		},
		{
			"DELETE",
			"/johnsmith/photos/puppy.jpg",
			map[string]string{
				"Host":       "s3.amazonaws.com",
				"Date":       "Tue, 27 Mar 2007 21:20:27 +0000",
				"User-Agent": "dotnet",
				"x-amz-date": "Tue, 27 Mar 2007 21:20:26 +0000"},
			map[string]string{},
			"k3nL7gH3+PadhTEVn5Ip83xlYzk=",
		},
		{
			"PUT",
			"/static.johnsmith.net/db-backup.dat.gz",
			map[string]string{
				"Host":                         "static.johnsmith.net:8080",
				"Date":                         "Tue, 27 Mar 2007 21:06:08 +0000",
				"User-Agent":                   "curl/7.15.5",
				"x-amz-acl":                    "public-read",
				"content-type":                 "application/x-download",
				"Content-MD5":                  "4gJE4saaMU4BqNR0kLY+lw==",
				"X-Amz-Meta-ReviewedBy":        "joe@johnsmith.net,jane@johnsmith.net",
				"X-Amz-Meta-FileChecksum":      "0x02661779",
				"X-Amz-Meta-ChecksumAlgorithm": "crc32",
				"Content-Disposition":          "attachment; filename=database.dat",
				"Content-Encoding":             "gzip",
				"Content-Length":               "5913339"},
			map[string]string{},
			"C0FlOtU8Ylb9KDTpZqYkZPX91iI=",
		},
		{
			"GET",
			"/dictionary/fran%C3%A7ais/pr%c3%a9f%c3%a8re",
			map[string]string{
				"Host": "s3.amazonaws.com",
				"Date": "Wed, 28 Mar 2007 01:49:49 +0000"},
			map[string]string{},
			"dxhSBHoI6eVSPcXJqEghlUzZMnY=",
		},
	}

	for i, test := range testCases {
		v := url.Values{}
		for key, p := range test.Params {
			v.Set(key, p)
		}
		var path string = test.Path
		if v.Encode() != "" {
			path = test.Path + "?" + v.Encode()
		}
		request, err := http.NewRequest(test.Method, path, nil)
		if err != nil {
			t.Errorf("[%d] unexcepted error %v", i, err)
		}
		for key, header := range test.Headers {
			request.Header.Add(key, header)
		}
		var signature string
		signature, err = GetSign(AccessKey, SecretKey, request)
		if err != nil {
			t.Errorf("[%d] unexcepted error %v", i, err)
		}
		if signature != test.Signature {
			t.Errorf("[%d] excepted signature %v, but got %v", i, test.Signature, signature)
		}
	}
}
