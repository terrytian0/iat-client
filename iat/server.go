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
)

var server = "http://127.0.0.1:8080/task/get"




func UploadTaskResult(taskResult TaskResult) (bool, error) {
	url := "http://127.0.0.1:8080/task/result/task"
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
	tr.Client = "127.0.0.1"
	tr.Id = id
	tr.StartTime = startTime
	tr.EndTime = endTime
	tr.Status = status
	return *tr

}


func UploadParameterResult(parameterResult ParameterResult) (bool, error) {
	url := "http://127.0.0.1:8080/task/result/parameter"
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
	pr.Client = "127.0.0.1"
	pr.Id = id
	pr.StartTime = startTime
	pr.EndTime = endTime
	pr.Status = strconv.FormatBool(status)
	pr.Message = message
	return *pr

}

func UploadApiResult(apiResult ApiResult) (bool, error) {
	url := "http://127.0.0.1:8080/task/result/api"
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


func GetApiResult(api Api, parameterId int64, requestHeaders map[string]string, requestBody string, header http.Header, body string, extractors []Extractor, asserts []Assert, startTime int64, endTime int64, status bool, message string) ApiResult {
	var ar = new(ApiResult)
	ar.Client = "127.0.0.1"
	ar.TaskId = api.TaskId
	ar.TestplanId = api.TestplanId
	ar.TestcaseId = api.TestcaseId
	ar.ParameterId = parameterId
	ar.TestcaseKeywordId = api.TestcaseKeywordId
	ar.KeywordId = api.KeywordId
	ar.KeywordApiId = api.KeywordApiId
	ar.ApiId = api.ApiId
	ar.Status = strconv.FormatBool(status)
	rh, _ := json.Marshal(requestHeaders)
	ar.RequestHeaders = string(rh)
	ar.RequestBody = requestBody
	ar.ResponseBody = body
	responseHeader, _ := json.Marshal(header)
	ar.ResponseHeaders = string(responseHeader)
	er, _ := json.Marshal(extractors)
	ar.Extractors = string(er)
	as, _ := json.Marshal(asserts)
	ar.Asserts = string(as)
	ar.StartTime = startTime
	ar.EndTime = endTime
	ar.Message = message
	return *ar
}

func GetTask() (*Task, error) {
	response, err := http.Get(server)
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

func GetUrl(parameter map[string]string, api Api) string {
	u, _ := url.Parse("http://127.0.0.1:8080" + api.Path)
	q := u.Query()
	form := GetFormData(parameter, api.Formdatas)
	if form != nil {
		for name, value := range form {
			q.Set(name, value)
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
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