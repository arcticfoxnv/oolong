package wirelesstag

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "strings"
    "time"
)

var apiHost = "www.mytaglist.com"

// Various API modules - docs for each are at http://wirelesstag.net/media/mytaglist.com/apidoc.html
var (
    ethAccount   = "ethAccount.asmx"
    ethClient    = "ethClient.asmx"
    ethLogs      = "ethLogs.asmx"
    ethLogShared = "ethLogShared.asmx"
)

// Client represents an API client.  It's not limited to the ethClient module
type Client interface {
    // GetMultiTagStatsRaw calls a method of the same name in the ethLogs module
    GetMultiTagStatsRaw([]int, string, time.Time, time.Time) ([]RawMultiStat, error)

    // GetStatsRaw calls a method of the same name in the ethLogs module
    GetStatsRaw(int, time.Time, time.Time) ([]RawStat, error)

    // GetTagManagers calls a method of the same name in the ethAccount module
    GetTagManagers() ([]TagManager, error)

    // GetTagManagerTagList calls a method of the same name in the ethClient module
    GetTagManagerTagList() (map[string][]Tag, error)
}

type clientError struct {
    Error error
    StatusCode int
    RequestURI string
    ResponseBody []byte
}

type wirelessTagClient struct {
    AccessToken string
}

type tagList []Tag
type tagManagerTagList struct {
    Mac string
    Tags tagList
}

type multiTagRawResponse struct {
    Stats []RawMultiStat
}

func NewClient(accessToken string) Client {
    return &wirelessTagClient{
        AccessToken: accessToken,
    }
}

// doPostEmptyRequest is a helper function for calling endpoints that take no input
func (c *wirelessTagClient) doPostEmptyRequest(module, endpoint string) ([]byte, *clientError) {
    content := strings.NewReader("{}")
    return c.doPostRequest(module, endpoint, content)
}

// doPostRequest makes a POST to the API server, and handles adding the authorization and content-type headers
func (c *wirelessTagClient) doPostRequest(module, endpoint string, content io.Reader) ([]byte, *clientError) {
    httpClient := &http.Client{}
    authStr := fmt.Sprintf("Bearer %s", c.AccessToken)
    url := fmt.Sprintf("https://%s/%s/%s", apiHost, module, endpoint)

    // Build the request
    req, err := http.NewRequest(http.MethodPost, url, content)
    req.Header.Add("Authorization", authStr)
    req.Header.Add("Content-Type", "application/json")
    if err != nil { return nil, &clientError{Error: err} }

    // Execute
    resp, err := httpClient.Do(req)
    if err != nil { return nil, &clientError{Error: err, RequestURI: url} }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, &clientError{Error: err, RequestURI: url, StatusCode: resp.StatusCode}
    }
    if resp.StatusCode != http.StatusOK {
        return nil, &clientError{Error: err, RequestURI: url, StatusCode: resp.StatusCode, ResponseBody: body}
    }

    return body, nil
}

func (c *wirelessTagClient) GetTagManagers() ([]TagManager, error) {
    resp, cErr := c.doPostEmptyRequest(ethAccount, "GetTagManagers")
    if cErr != nil { return nil, cErr.Error }

    decodedResponse := make(map[string][]TagManager, 0)
    err := json.Unmarshal(resp, &decodedResponse)
    if err != nil { return nil, err }
    return decodedResponse["d"], nil
}

func (c *wirelessTagClient) GetTagManagerTagList() (map[string][]Tag, error) {
    resp, cErr := c.doPostEmptyRequest(ethClient, "GetTagManagerTagList")
    if cErr != nil { return nil, cErr.Error }

    decodedResponse := make(map[string][]tagManagerTagList, 0)
    err := json.Unmarshal(resp, &decodedResponse)
    if err != nil { return nil, err }

    list := make(map[string][]Tag, 0)
    for _, entry := range decodedResponse["d"] {
        list[entry.Mac] = entry.Tags
    }

    return list, nil
}

func (c *wirelessTagClient) GetStatsRaw(slaveId int, from, to time.Time) ([]RawStat, error) {
    data, err := json.Marshal(map[string]interface{}{
        "id": slaveId,
        "fromDate": from.Format(DateFormat),
        "toDate": to.Format(DateFormat),
    })
    if err != nil { return nil, err }
    content := bytes.NewReader(data)

    resp, cErr := c.doPostRequest(ethLogs, "GetStatsRaw", content)
    if cErr != nil {
        return nil, cErr.Error
    }

    decodedResponse := make(map[string][]RawStat, 0)
    err = json.Unmarshal(resp, &decodedResponse)
    if err != nil { return nil, err }

    return decodedResponse["d"], nil
}

func (c *wirelessTagClient) GetMultiTagStatsRaw(slaveIds []int, statType string, from, to time.Time) ([]RawMultiStat, error) {
    data, err := json.Marshal(map[string]interface{}{
        "ids": slaveIds,
        "type": statType,
        "fromDate": from.Format(DateFormat),
        "toDate": to.Format(DateFormat),
    })
    if err != nil { return nil, err }
    content := bytes.NewReader(data)

    resp, cErr := c.doPostRequest(ethLogs, "GetMultiTagStatsRaw", content)
    if cErr != nil {
        return nil, cErr.Error
    }

    decodedResponse := make(map[string]multiTagRawResponse, 0)
    err = json.Unmarshal(resp, &decodedResponse)
    if err != nil { return nil, err }

    return decodedResponse["d"].Stats, nil
}
