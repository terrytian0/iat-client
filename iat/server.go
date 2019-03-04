package iat

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"errors"
	"net/url"
	"net"
	"time"
	"strings"
	"strconv"
	"os"
	"github.com/oliveagle/jsonpath"
)

var Server string
var Client string
var Key string

const (
	REGISTER         = "/client/register"
	HEARTBEAT        = "/client/heartbeat"
	GET_TASK         = "/task/get"
	RESULT_TASK      = "/task/result/task"
	RESULT_PARAMETER = "/task/result/parameter"
	RESULT_API       = "/task/result/api"
)

func Register() (bool, error) {
	url := "http://" + Server + REGISTER + "?client=" + Client
	response, err := http.Post(url, "application/json;charset=UTF-8", nil)
	if err != nil {
		return false, err
	}
	if response.StatusCode != 200 {
		return false, errors.New("http code " + strconv.Itoa(response.StatusCode))
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, err
	}

	var jsonData interface{}
	err = json.Unmarshal([]byte(body), &jsonData)
	if err != nil {
		fmt.Println("body not json")
		return false, err
	}
	status, err := jsonpath.JsonPathLookup(jsonData, "$.status")
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	if status.(bool) == false {
		fmt.Println(body)
		return false, err
	}
	key, err := jsonpath.JsonPathLookup(jsonData, "$.content.key")
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	Key = key.(string)
	return true, nil
}

func Heartbeat() {
	url := "http://" + Server + HEARTBEAT + "?client=" + Client + "&key=" + Key
	response, err := http.Post(url, "application/json;charset=UTF-8", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	if response.StatusCode != 200 {
		fmt.Println(errors.New("http code " + strconv.Itoa(response.StatusCode)))
		return
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	var result ServerResult
	json.Unmarshal(body, &result)
	if result.Status == false {
		fmt.Println(result.Message)
		return
	}
}

func UploadTaskResult(taskResult TaskResult) (bool, error) {
	url := "http://" + Server + RESULT_TASK
	rBody, err := json.Marshal(taskResult)
	fmt.Println(string(rBody))
	response, err := http.Post(url, "application/json;charset=UTF-8", strings.NewReader(string(rBody)))
	if err != nil {
		return false, err
	}
	if response.StatusCode != 200 {
		return false, errors.New("http code " + strconv.Itoa(response.StatusCode))
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, err
	}
	var result ServerResult
	json.Unmarshal(body, &result)
	if result.Status == false {
		return false, errors.New(result.Message)
	}
	return true, nil
}

func GetTaskResult(id int64, startTime int64, endTime int64, status string) TaskResult {
	var tr = new(TaskResult)
	tr.Client = Client
	tr.Id = id
	tr.StartTime = startTime
	tr.EndTime = endTime
	tr.Status = status
	tr.Key = Key
	return *tr

}

func UploadParameterResult(parameterResult ParameterResult) (bool, error) {
	url := "http://" + Server + RESULT_PARAMETER
	rBody, err := json.Marshal(parameterResult)
	fmt.Println(string(rBody))
	response, err := http.Post(url, "application/json;charset=UTF-8", strings.NewReader(string(rBody)))
	if err != nil {
		return false, err
	}
	if response.StatusCode != 200 {
		return false, errors.New("http code " + strconv.Itoa(response.StatusCode))
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, err
	}
	var result ServerResult
	json.Unmarshal(body, &result)
	if result.Status == false {
		return false, errors.New(result.Message)
	}
	return true, nil
}

func GetParameterResult(id int64, startTime int64, endTime int64, status bool, message string) ParameterResult {
	var pr = new(ParameterResult)
	pr.Client = Client
	pr.Id = id
	pr.StartTime = startTime
	pr.EndTime = endTime
	pr.Status = strconv.FormatBool(status)
	pr.Message = message
	pr.Key = Key
	return *pr

}

func UploadApiResult(apiResult ApiResult) (bool, error) {
	url := "http://" + Server + RESULT_API
	rBody, err := json.Marshal(apiResult)
	fmt.Println(string(rBody))
	response, err := http.Post(url, "application/json;charset=UTF-8", strings.NewReader(string(rBody)))
	if err != nil {
		return false, err
	}
	if response.StatusCode != 200 {
		return false, errors.New("http code " + strconv.Itoa(response.StatusCode))
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, err
	}
	var result ServerResult
	json.Unmarshal(body, &result)
	if result.Status == false {
		return false, errors.New(result.Message)
	}
	return true, nil
}

func GetApiResult(api Api, url string, parameterId int64, requestHeaders map[string]string, requestBody string, header http.Header, responseBody []byte, extractors []Extractor, asserts []Assert, startTime int64, endTime int64, status bool, message string) ApiResult {
	var ar = new(ApiResult)
	ar.Client = Client
	ar.TaskId = api.TaskId
	ar.TestplanId = api.TestplanId
	ar.TestcaseId = api.TestcaseId
	ar.ParameterId = parameterId
	ar.TestcaseKeywordId = api.TestcaseKeywordId
	ar.KeywordId = api.KeywordId
	ar.KeywordApiId = api.KeywordApiId
	ar.ApiId = api.ApiId
	ar.Url = url
	ar.Method = api.Method
	ar.Status = strconv.FormatBool(status)
	if requestHeaders != nil {
		rh, _ := json.Marshal(requestHeaders)
		ar.RequestHeaders = string(rh)
	}
	ar.RequestBody = requestBody
	if responseBody != nil {
		ar.ResponseBody = string(responseBody)
	}
	if header != nil {
		responseHeader, _ := json.Marshal(header)
		ar.ResponseHeaders = string(responseHeader)
	}
	if extractors != nil {
		er, _ := json.Marshal(extractors)
		ar.Extractors = string(er)
	}
	if asserts != nil {
		as, _ := json.Marshal(asserts)
		ar.Asserts = string(as)
	}
	ar.StartTime = startTime
	ar.EndTime = endTime
	ar.Message = message
	ar.Key = Key
	return *ar
}

func GetTask() (*Task, error) {
	url := "http://" + Server + GET_TASK + "?client=" + Client + "&key=" + Key
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		fmt.Println(response.StatusCode)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var result ServerResult
	json.Unmarshal(body, &result)
	if result.Status == false {
		return nil, errors.New(result.Message)
	}
	return &result.Content, nil
}

func GetUrl(parameter map[string]string, api Api, envs map[int64]string) (string, error) {
	//TODO ENV需要带服务器地址
	host := envs[api.ServiceId]
	if host == "" {
		return "", errors.New("服务为设置环境变量！")
	}
	u, _ := url.Parse("http://" + host + api.Path)
	q := u.Query()
	form := GetFormData(parameter, api.Formdatas)
	if form != nil {
		for name, value := range form {
			q.Set(name, value)
		}
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func GetRequest(url string, method string, headers map[string]string, body string) *http.Request {
	req, err := http.NewRequest(strings.ToUpper(method), url, strings.NewReader(body))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	return req
}

func GetClient() (*http.Client, error) {
	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(25 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*20)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}
	return client, nil
}

func GetTimestamp() int64 {
	return int64(time.Now().UnixNano() / (1000 * 1000))
}

func GetLocalIp() string {
	addr, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, address := range addr {
		if ip, ok := address.(*net.IPNet); ok && !ip.IP.IsLoopback() {
			if ip.IP.To4() != nil {
				return ip.IP.String()
			}
		}
	}
	return ""
}
