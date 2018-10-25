package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

func TestWorkflow(t *testing.T) {
	inst, err := aetest.NewInstance(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer inst.Close()

	ctx1, res1, req1 := request(http.MethodGet, "/root", nil, inst, t)
	getOrDelete(ctx1, "root", res1, req1)

	if res1.Code != 500 {
		t.Errorf("Expected Response to be %v, got %v", 500, res1.Code)
	}

	task := Task{ID: "root", Text: "b", Checked: false, Children: []string{"d"}}
	data, err := json.Marshal(task)
	if err != nil {
		t.Fatal(err)
	}

	reader := bytes.NewReader(data)

	ctx2, res2, req2 := request(http.MethodPost, "/set", reader, inst, t)
	set(ctx2, res2, req2)

	if res2.Code != 200 {
		t.Errorf("Expected Response to be %v, got %v", 200, res2.Code)
	}

	ctx3, res3, req3 := request(http.MethodGet, "/root", nil, inst, t)
	getOrDelete(ctx3, "root", res3, req3)

	if res3.Code != 200 {
		t.Errorf("Expected Response to be %v, got %v", 200, res3.Code)
	}

	ctx4, res4, req4 := request(http.MethodDelete, "/root", nil, inst, t)
	getOrDelete(ctx4, "root", res4, req4)

	if res4.Code != 200 {
		t.Errorf("Expected Response to be %v, got %v", 200, res4.Code)
	}

	ctx5, res5, req5 := request(http.MethodGet, "/root", nil, inst, t)
	getOrDelete(ctx5, "root", res5, req5)

	if res5.Code != 500 {
		t.Errorf("Expected Response to be %v, got %v", 500, res5.Code)
	}
}

func request(
	method string,
	path string,
	body io.Reader,
	inst aetest.Instance,
	t *testing.T,
) (context.Context, *httptest.ResponseRecorder, *http.Request) {
	req, err := inst.NewRequest(method, path, body)
	if err != nil {
		t.Fatal(err)
	}

	ctx, res := appengine.NewContext(req), httptest.NewRecorder()

	return ctx, res, req
}
