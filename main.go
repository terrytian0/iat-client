package main

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"github.com/pkg/errors"
	"time"
	"net"
	"strings"
	"regexp"
	"github.com/oliveagle/jsonpath"
	"strconv"
	"net/url"
	"sync"
)

var server = "http://127.0.0.1:8080/task/get"
var currentTask = make(map[int64]Task)
var mutex sync.Mutex

func main() {
	for true {
		if len(currentTask) < 3 {
			go exec()
		}
		fmt.Println("当前"+strconv.Itoa(len(currentTask))+"个任务正在执行，sleep 10 secend!")
		time.Sleep(1 * time.Second)
	}
}





func exec() {
	task, err := getTask()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer delete(currentTask, task.Id)
	currentTask[task.Id] = *task
	startTime := getTimestamp()
	time.Sleep(10*time.Second)
	runTask(*task)
	endTime := getTimestamp()
	taskResult := getTaskResult(task.Id, startTime, endTime, "FINISHED")
	uploadTaskResult(taskResult)
}


func getTaskResult(id int64, startTime int64, endTime int64, status string) TaskResult {
	var tr = new(TaskResult)
	tr.Client = "127.0.0.1"
	tr.Id = id
	tr.StartTime = startTime
	tr.EndTime = endTime
	tr.Status = status
	return *tr

}

func getTask() (*Task, error) {
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

func runTask(task Task) {
	for _, testcase := range task.Testcases {
		runTestcase(testcase)
	}
}

func runTestcase(testcase Testcase) (bool, string) {
	for _, parameter := range testcase.Parameters {
		p := make(map[string]string)
		p = getParameter(p, parameter)
		startTime := getTimestamp()
		p, res, message := runParameter(p, parameter.Id, testcase.Keywords)
		endTime := getTimestamp()
		parameterResult := getParameterResult(parameter.Id, startTime, endTime, res, message)
		uploadParameterResult(parameterResult)
		if res == false {
			return false, message
		}
	}
	return true, ""
}

func getParameterResult(id int64, startTime int64, endTime int64, status bool, message string) ParameterResult {
	var pr = new(ParameterResult)
	pr.Client = "127.0.0.1"
	pr.Id = id
	pr.StartTime = startTime
	pr.EndTime = endTime
	pr.Status = strconv.FormatBool(status)
	pr.Message = message
	return *pr

}
func runParameter(parameter map[string]string, parameterId int64, keywords []Keyword) (map[string]string, bool, string) {
	for _, keyword := range keywords {
		parameter, res, message := runKeyword(parameter, parameterId, keyword)
		if res == false {
			return parameter, false, message
		}
	}
	return parameter, true, ""
}
func runKeyword(parameter map[string]string, parameterId int64, keyword Keyword) (map[string]string, bool, string) {
	res := true
	for _, api := range keyword.Apis {
		parameter, res, message := runApi(parameter, parameterId, api)
		if res == false {
			return parameter, res, message
		}
	}
	return parameter, res, ""
}
func runApi(parameter map[string]string, parameterId int64, api Api) (map[string]string, bool, string) {
	client, err := getClient()
	if err != nil {
		fmt.Println(err)
	}
	u := getUrl(parameter, api)
	requestBody := getBody(parameter, api.Body)
	requestHeaders := getHeader(parameter, api.Headers)
	req := getRequest(u, api.Method, requestHeaders, requestBody)
	startTime := getTimestamp()
	response, err := client.Do(req)
	endTime := getTimestamp()
	defer response.Body.Close()
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}
	extractors := getExtractor(api.Extractors)
	parameter, extractors = extractor(parameter, string(responseBody), extractors)
	asserts := getAssert(api.Asserts)
	res, asserts, message := assert(parameter, response.StatusCode, string(responseBody), response.Header, asserts)
	apiResult := getApiResult(api, parameterId, requestHeaders, string(responseBody), response.Header, string(responseBody), extractors, asserts, startTime, endTime, res, message)
	uploadApiResult(apiResult)
	return parameter, res, message
}

func getApiResult(api Api, parameterId int64, requestHeaders map[string]string, requestBody string, header http.Header, body string, extractors []Extractor, asserts []Assert, startTime int64, endTime int64, status bool, message string) ApiResult {
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

func uploadTaskResult(taskResult TaskResult) (bool, error) {
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

func uploadParameterResult(parameterResult ParameterResult) (bool, error) {
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

func uploadApiResult(apiResult ApiResult) (bool, error) {
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

func extractor(parameter map[string]string, body string, es []Extractor) (map[string]string, []Extractor) {
	for i, e := range es {
		if e.Type == "JSON" {
			var jsonData interface{}
			err := json.Unmarshal([]byte(body), &jsonData)
			if err != nil {
				fmt.Println("body not json")
				return parameter, es
			}
			res, err := jsonpath.JsonPathLookup(jsonData, e.Rule)
			if err != nil {
				fmt.Println(err)
				continue
			}
			value := interfaceToString(res)
			parameter[e.Name] = value
			e.Value = value
			es[i] = e
		} else if e.Type == "REGEXP" {

		} else {
			fmt.Println("extractor type error!")
		}
	}
	return parameter, es
}

func interfaceToString(res interface{}) string {
	switch res.(type) {
	case int:
		v := res.(int)
		return strconv.Itoa(v)
	case float64:
		v := res.(float64)
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return res.(string)
	default:
		fmt.Println("json type error")
		return res.(string)
	}
}

func assert(parameter map[string]string, httpcode int, body string, header http.Header, asserts []Assert) (bool, []Assert, string) {
	if asserts == nil {
		return true, asserts, ""
	}
	for i, a := range asserts {
		value := ""
		if a.Locale == "HTTPCODE" {
			value = strconv.Itoa(httpcode)
		} else if a.Locale == "HEADER" {
			value = header.Get(a.Rule)
		} else if a.Locale == "BODY" {
			var jsonData interface{}
			err := json.Unmarshal([]byte(body), &jsonData)
			if err != nil {
				fmt.Println("body not json")
				return false, asserts, ""
			}
			res, err := jsonpath.JsonPathLookup(jsonData, a.Rule)
			value = interfaceToString(res)
		} else {
			fmt.Println("assert type error!")
		}
		cr, msg := compare(parameter, a.Method, a.Value, value)
		if cr {
			a.Status = strconv.FormatBool(true)
			asserts[i] = a
			continue
		} else {
			a.Status = strconv.FormatBool(false)
			asserts[i] = a
			return false, asserts, msg
		}

	}
	return true, asserts, ""
}

func compare(parameter map[string]string, method string, expect string, actual string) (bool, string) {
	e := parameterReplace(parameter, expect)
	if method == "CONTAINS" {
		if strings.Contains(actual, parameterReplace(parameter, expect)) {
			return true, ""
		} else {
			return false, actual + " no contains " + e
		}
	} else if method == "EQUALS" {
		if actual == parameterReplace(parameter, expect) {
			return true, ""
		} else {
			return false, actual + " no equals " + e
		}
	} else if method == "GREATER" {
		//TODO 待实现
		return true, ""
	} else if method == "LESS" {
		//TODO 待实现
		return true, ""
	} else {
		return true, ""
	}
}

func getParameter(p map[string]string, parameter Parameter) map[string]string {
	if parameter.Parameters == "" {
		return p
	}
	pp := make(map[string]string)
	err := json.Unmarshal([]byte(parameter.Parameters), &pp)
	if err != nil {
		fmt.Println(err)
	}
	for k, v := range pp {
		p[k] = v
	}
	return p
}

func getRequest(url string, method string, headers map[string]string, body string) *http.Request {
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

func getUrl(parameter map[string]string, api Api) string {
	u, _ := url.Parse("http://127.0.0.1:8080" + api.Path)
	q := u.Query()
	form := getFormData(parameter, api.Formdatas)
	if form != nil {
		for name, value := range form {
			q.Set(name, value)
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func getExtractor(extractor string) []Extractor {
	if extractor == "" {
		return nil
	}
	var extractors []Extractor
	err := json.Unmarshal([]byte(extractor), &extractors)
	if err != nil {
		return nil
	}
	return extractors
}

func getAssert(assert string) []Assert {
	if assert == "" {
		return nil
	}
	var asserts []Assert
	err := json.Unmarshal([]byte(assert), &asserts)
	if err != nil {
		return nil
	}
	return asserts
}

func getHeader(parameter map[string]string, header string) map[string]string {
	if header == "" {
		return nil
	}
	var headers []Header
	err := json.Unmarshal([]byte(header), &headers)
	if err != nil {
		return nil
	}
	hds := make(map[string]string)
	for _, h := range headers {
		hds[h.Name] = parameterReplace(parameter, h.DefaultValue)
	}
	return hds
}

func getBody(parameter map[string]string, body string) string {
	if body == "" {
		return ""
	}
	var b Body
	err := json.Unmarshal([]byte(body), &b)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return parameterReplace(parameter, b.DefaultValue)
}

func getFormData(parameter map[string]string, formData string) map[string]string {
	if formData == "" {
		return nil
	}
	var formDatas []FormData
	err := json.Unmarshal([]byte(formData), &formDatas)
	if err != nil {
		return nil
	}
	form := make(map[string]string)
	for _, f := range formDatas {
		form[f.Name] = parameterReplace(parameter, f.DefaultValue)
	}
	return form
}

func parameterReplace(parameter map[string]string, str string) string {
	re := regexp.MustCompile("#{(.*?)}+")
	sub := re.FindAllStringSubmatch(str, -1)
	for _, a := range sub {
		//fmt.Println(a)
		if parameter[a[1]] != "" {
			str = strings.Replace(str, a[0], parameter[a[1]], -1)
		}
	}
	return str
}

func getClient() (*http.Client, error) {
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

func getTimestamp() int64 {
	return int64(time.Now().UnixNano() / (1000 * 1000))
}

type ApiResult struct {
	TaskId            int64  `json:"taskId"`
	TestplanId        int64  `json:"testplanId"`
	TestcaseId        int64  `json:"testcaseId"`
	ParameterId       int64  `json:"parameterId"`
	TestcaseKeywordId int64  `json:"testcaseKeywordId"`
	KeywordId         int64  `json:"keywordId"`
	KeywordApiId      int64  `json:"keywordApiId"`
	ApiId             int64  `json:"apiId"`
	RequestHeaders    string `json:"requestHeaders"`
	RequestFormdatas  string `json:"requestFormdatas"`
	RequestBody       string `json:"requestBody"`
	ResponseHeaders   string `json:"responseHeaders"`
	ResponseBody      string `json:"responseBody"`
	Extractors        string `json:"extractors"`
	Asserts           string `json:"asserts"`
	Status            string `json:"status"`
	Client            string `json:"client"`
	Key               string `json:"key"`
	Message           string `json:"message"`
	StartTime         int64  `json:"startTime"`
	EndTime           int64  `json:"endTime"`
}

type ParameterResult struct {
	Id        int64  `json:"id"`
	Status    string `json:"status"`
	Client    string `json:"client"`
	Key       string `json:"key"`
	Message   string `json:"message"`
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
}

type TaskResult struct {
	Id        int64  `json:"id"`
	Status    string `json:"status"`
	Client    string `json:"client"`
	Key       string `json:"key"`
	Message   string `json:"message"`
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
}

type ServerResult struct {
	Status  bool   `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Content Task   `json:"content"`
}

type Extractor struct {
	Id           int64  `json:"id"`
	KeywordApiId int64  `json:"keywordApiId"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	Rule         string `json:"rule"`
	Description  string `json:"description"`
	Value        string `json:"value"`
}

type Assert struct {
	Id           int64  `json:"id"`
	KeywordApiId int64  `json:"keywordApiId"`
	Type         string `json:"type"`
	Locale       string `json:"locale"`
	Rule         string `json:"rule"`
	Method       string `json:"method"`
	Value        string `json:"value"`
	Description  string `json:"description"`
	Status       string `json:"status"`
}

type FormData struct {
	ApiId        int64  `json:"apiId"`
	DefaultValue string `json:"defaultValue"`
	Id           int64  `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
}

type Header struct {
	ApiId        int64  `json:"apiId"`
	DefaultValue string `json:"defaultValue"`
	Id           int64  `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
}

type Body struct {
	ApiId        int64  `json:"apiId"`
	DefaultValue string `json:"defaultValue"`
	Id           int64  `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
}

type Api struct {
	Id                int64  `json:"id"`
	TaskId            int64  `json:"taskId"`
	TestplanId        int64  `json:"testplanId"`
	TestcaseId        int64  `json:"testcaseId"`
	TestcaseKeywordId int64  `json:"testcaseKeywordId"`
	KeywordId         int64  `json:"keywordId"`
	KeywordApiId      int64  `json:"keywordApiId"`
	ApiId             int64  `json:"apiId"`
	Path              string `json:"path"`
	Method            string `json:"method"`
	Version           int32  `json:"version"`
	Idx               int32  `json:"idx"`
	Headers           string `json:"headers"`
	Formdatas         string `json:"formdatas"`
	Body              string `json:"body"`
	Extractors        string `json:"extractors"`
	Asserts           string `json:"asserts"`
}

type Keyword struct {
	Id                int64  `json:"id"`
	TaskId            int64  `json:"taskId"`
	TestcaseId        int64  `json:"testcaseId"`
	TestcaseKeywordId int64  `json:"testcaseKeywordId"`
	KeywordId         int64  `json:"keywordId"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	Idx               int32  `json:"idx"`
	Apis              []Api  `json:"apis"`
}

type Parameter struct {
	Id         int64  `json:"id"`
	TaskId     int64  `json:"taskId"`
	TestplanId int64  `json:"testplanId"`
	TestcaseId int64  `json:"testcaseId"`
	Parameters string `json:"parameters"`
	Status     string `json:"status"`
}
type Testcase struct {
	Id          int64       `json:"id"`
	TaskId      int64       `json:"taskId"`
	TestcaseId  int64       `json:"testcaseId"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Idx         int32       `json:"idx"`
	Status      string      `json:"status"`
	Keywords    []Keyword   `json:"keywords"`
	Parameters  []Parameter `json:"parameters"`
}

type Task struct {
	Id           int64  `json:"id"`
	ServiceId    int64  `json:"serviceId"`
	TestplanId   int64  `json:"testplanId"`
	TestplanName string `json:"testplanName"`
	PassRate     int32  `json:"passRate"`
	Coverage     int32  `json:"coverage"`
	Status       string `json:"status"`
	Client       string `json:"client"`
	CreateUser   string `json:"createUser"`
	//CreateTime   time.Time `json:"createTime"`
	UpdateUser string `json:"updateUser"`
	//UpdateTime   time.Time `json:"updateTime"`
	Testcases []Testcase `json:"testcases"`
}
