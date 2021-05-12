package onfido

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
)

func TestNewWebhookFromEnv_MissingToken(t *testing.T) {
	_, err := NewWebhookFromEnv()
	if err == nil {
		t.Fatal()
	}
	if err != ErrMissingWebhookToken {
		t.Fatal("expected error to match ErrMissingWebhookToken")
	}
}

func TestNewWebhookFromEnv_TokenSet(t *testing.T) {
	expected := "808yup"
	os.Setenv(WebhookTokenEnv, expected)
	defer os.Setenv(WebhookTokenEnv, "")

	wh, err := NewWebhookFromEnv()
	if err != nil {
		t.Fatal()
	}
	if wh.(*webhook).Token != expected {
		t.Fatalf("expected to see `%s` token but got `%s`", expected, wh.(*webhook).Token)
	}
}

func TestValidateSignature_InvalidSignature(t *testing.T) {
	wh := webhook{Token: "abc123"}
	err := wh.ValidateSignature([]byte("hello world"), "invalid")
	if err == nil {
		t.Fatal()
	}
	if err != ErrInvalidWebhookSignature {
		t.Fatal("expected error to match ErrInvalidWebhookSignature")
	}
}

func TestValidateSignature_ValidSignature(t *testing.T) {
	wh := webhook{Token: "abc123"}
	err := wh.ValidateSignature([]byte("hello world"), "8c301acf7e955038b486de8f2a35f7f28bb5755fd1f77e1dbf9ef9e27713ad0d")
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseFromRequest_InvalidSignature(t *testing.T) {
	req := &http.Request{
		Header: make(map[string][]string),
	}
	req.Header.Add(WebhookSignatureHeader, "123")
	req.Body = ioutil.NopCloser(bytes.NewBuffer([]byte("{\"msg\": \"hello world\"}")))

	wh := webhook{Token: "abc123"}
	_, err := wh.ParseFromRequest(req)
	if err == nil {
		t.Fatal()
	}
	if err != ErrInvalidWebhookSignature {
		t.Fatal("expected error to match ErrInvalidWebhookSignature")
	}
}

func TestParseFromRequest_SkipSignatureValidation(t *testing.T) {
	req := &http.Request{
		Header: make(map[string][]string),
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer([]byte("{\"msg\": \"hello world\"}")))

	wh := webhook{Token: "abc123", SkipSignatureValidation: true}
	_, err := wh.ParseFromRequest(req)
	if err != nil {
		t.Errorf("expected no error as signature validation should have been skipped: %s", err.Error())
	}
}

func TestParseFromRequest_InvalidJson(t *testing.T) {
	req := &http.Request{
		Header: make(map[string][]string),
	}
	req.Header.Add(WebhookSignatureHeader, "d4163f7af2256fae6ab72cb595d3f9d1dfc6fecc")
	req.Body = ioutil.NopCloser(bytes.NewBuffer([]byte("{\"msg\": \"hello world")))

	wh := webhook{Token: "abc123"}
	_, err := wh.ParseFromRequest(req)
	if err == nil {
		t.Fatal("expected invalid json to raise an error")
	}
}

func TestParseFromRequest_ValidSignature(t *testing.T) {
	req := &http.Request{
		Header: make(map[string][]string),
	}
	req.Header.Add(WebhookSignatureHeader, "b469eabb36776543320fc09ed03451c34706daa3a730a561868ab2cc4399f8ec")
	req.Body = ioutil.NopCloser(bytes.NewBuffer([]byte("{\"msg\": \"hello world\"}")))

	wh := webhook{Token: "abc123"}
	_, err := wh.ParseFromRequest(req)
	if err != nil {
		t.Fatal(err)
	}
}
