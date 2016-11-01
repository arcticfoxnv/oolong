package wirelesstag

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDoPostRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `you win!`)
	}))
	defer ts.Close()
	apiHost = ts.URL

	client := &wirelessTagClient{AccessToken: "xyz"}
	content := strings.NewReader("")
	res, err := client.doPostRequest(ethAccount, "testing", content)
	if err != nil {
		t.Fail()
	}
	if len(res) != 8 {
		t.Fail()
	}
}

func TestDoPostRequestBadAuth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `you win!`)
	}))
	defer ts.Close()
	apiHost = ts.URL

	client := &wirelessTagClient{AccessToken: "xyz"}
	content := strings.NewReader("")
	res, err := client.doPostRequest(ethAccount, "testing", content)
	if err == nil {
		t.Fail()
	}
	if res != nil {
		t.Fail()
	}
}

func TestDoPostEmptyRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `you win!`)
	}))
	defer ts.Close()
	apiHost = ts.URL

	client := &wirelessTagClient{AccessToken: "xyz"}
	res, err := client.doPostEmptyRequest(ethAccount, "testing")
	if err != nil {
		t.Fail()
	}
	if len(res) != 8 {
		t.Fail()
	}
}

func TestGetTagManagers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"d": [{"Name": "Test Manager"}]}`)
	}))
	defer ts.Close()
	apiHost = ts.URL

	client := NewClient("xyz")
	res, err := client.GetTagManagers()
	if err != nil {
		t.Fail()
	}
	if res[0].Name != "Test Manager" {
		t.Fail()
	}
}

func TestGetTagManagersBadResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `garbage`)
	}))
	defer ts.Close()
	apiHost = ts.URL

	client := NewClient("xyz")
	res, err := client.GetTagManagers()
	if err == nil {
		t.Fail()
	}
	if res != nil {
		t.Fail()
	}
}

func TestGetTagManagersBadData(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `garbage`)
	}))
	defer ts.Close()
	apiHost = ts.URL

	client := NewClient("xyz")
	res, err := client.GetTagManagers()
	if err == nil {
		t.Fail()
	}
	if res != nil {
		t.Fail()
	}
}

func TestGetTagManagerTagList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"d": [{"mac": "xx:xx:xx:xx:xx:xx", "tags": [{"Name": "Test Tag"}]}]}`)
	}))
	defer ts.Close()
	apiHost = ts.URL

	client := NewClient("xyz")
	res, err := client.GetTagManagerTagList()
	if err != nil {
		t.Fail()
	}
	if res["xx:xx:xx:xx:xx:xx"][0].Name != "Test Tag" {
		t.Fail()
	}
}

func TestGetTagManagerTagListBadResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `garbage`)
	}))
	defer ts.Close()
	apiHost = ts.URL

	client := NewClient("xyz")
	res, err := client.GetTagManagerTagList()
	if err == nil {
		t.Fail()
	}
	if res != nil {
		t.Fail()
	}
}

func TestGetTagManagerTagListBadData(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `garbage`)
	}))
	defer ts.Close()
	apiHost = ts.URL

	client := NewClient("xyz")
	res, err := client.GetTagManagerTagList()
	if err == nil {
		t.Fail()
	}
	if res != nil {
		t.Fail()
	}
}

func TestGetStatsRaw(t *testing.T) {
	now := time.Now()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"d": [{"date": "%s"}]}`, now.Format(DateFormat))
	}))
	defer ts.Close()
	apiHost = ts.URL

	client := NewClient("xyz")
	res, err := client.GetStatsRaw(0, now, now)
	if err != nil {
		t.Fail()
	}
	if res[0].Date != now.Format(DateFormat) {
		t.Fail()
	}
}

func TestGetStatsRawBadResponse(t *testing.T) {
	now := time.Now()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `garbage`)
	}))
	defer ts.Close()
	apiHost = ts.URL

	client := NewClient("xyz")
	res, err := client.GetStatsRaw(0, now, now)
	if err == nil {
		t.Fail()
	}
	if res != nil {
		t.Fail()
	}
}

func TestGetStatsRawBadData(t *testing.T) {
	now := time.Now()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `garbage`)
	}))
	defer ts.Close()
	apiHost = ts.URL

	client := NewClient("xyz")
	res, err := client.GetStatsRaw(0, now, now)
	if err == nil {
		t.Fail()
	}
	if res != nil {
		t.Fail()
	}
}

func TestGetMultiTagStatsRaw(t *testing.T) {
	now := time.Now()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"d": {"stats": [{"date": "%s", "ids": [0, 1], "values": [[0.0], [0.0]]}]}}`, now.Format(DateFormat))
	}))
	defer ts.Close()
	apiHost = ts.URL

	client := NewClient("xyz")
	res, err := client.GetMultiTagStatsRaw([]int{0, 1}, "something", now, now)
	if err != nil {
		t.Fail()
	}
	if res[0].Date != now.Format(DateFormat) {
		t.Fail()
	}
}

func TestGetMultiTagStatsRawBadResponse(t *testing.T) {
	now := time.Now()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `garbage`)
	}))
	defer ts.Close()
	apiHost = ts.URL

	client := NewClient("xyz")
	res, err := client.GetMultiTagStatsRaw([]int{0, 1}, "something", now, now)
	if err == nil {
		t.Fail()
	}
	if res != nil {
		t.Fail()
	}
}

func TestGetMultiTagStatsRawBadData(t *testing.T) {
	now := time.Now()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `garbage`)
	}))
	defer ts.Close()
	apiHost = ts.URL

	client := NewClient("xyz")
	res, err := client.GetMultiTagStatsRaw([]int{0, 1}, "something", now, now)
	if err == nil {
		t.Fail()
	}
	if res != nil {
		t.Fail()
	}
}
