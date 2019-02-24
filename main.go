package main

import (
	"fmt"
	"io/ioutil"
	"encoding/json"
	"time"
	"strings"
	"github.com/oliveagle/jsonpath"
	"strconv"
	"sync"
	"github.com/terrtian0/iat-client/iat"
	"net/http"
	"flag"
)

var currentTask = make(map[int64]iat.Task)
var mutex sync.Mutex

func main() {
	flag.StringVar(&iat.Server, "s", "127.0.0.1:8080", "iat server")
	flag.StringVar(&iat.Client, "l", "", "iat server")
	flag.Parse()
	if iat.Client == "" {
		iat.Client = iat.GetLocalIp()
	}
	res, err := iat.Register()
	if res == false {
		fmt.Println(err)
		return
	}
	go heartbeat()
	for true {
		if len(currentTask) < 3 {
			go exec()
		}
		fmt.Println("当前" + strconv.Itoa(len(currentTask)) + "个任务正在执行，sleep 10 secend!")
		time.Sleep(10 * time.Second)
	}
}

func heartbeat()  {
	for true  {
		iat.Heartbeat()
		time.Sleep(60 * time.Second)
	}
}

func exec() {
	task, err := iat.GetTask()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer delete(currentTask, task.Id)
	currentTask[task.Id] = *task
	startTime := iat.GetTimestamp()
	runTask(*task)
	endTime := iat.GetTimestamp()
	taskResult := iat.GetTaskResult(task.Id, startTime, endTime, "FINISHED")
	iat.UploadTaskResult(taskResult)
}

func runTask(task iat.Task) {
	for _, testcase := range task.Testcases {
		runTestcase(testcase)
	}
}

func runTestcase(testcase iat.Testcase) (bool, string) {
	for _, parameter := range testcase.Parameters {
		p := make(map[string]string)
		p = iat.GetParameter(p, parameter)
		startTime := iat.GetTimestamp()
		p, res, message := runParameter(p, parameter.Id, testcase.Keywords)
		endTime := iat.GetTimestamp()
		parameterResult := iat.GetParameterResult(parameter.Id, startTime, endTime, res, message)
		iat.UploadParameterResult(parameterResult)
		if res == false {
			return false, message
		}
	}
	return true, ""
}

func runParameter(parameter map[string]string, parameterId int64, keywords []iat.Keyword) (map[string]string, bool, string) {
	for _, keyword := range keywords {
		parameter, res, message := runKeyword(parameter, parameterId, keyword)
		if res == false {
			return parameter, false, message
		}
	}
	return parameter, true, ""
}

func runKeyword(parameter map[string]string, parameterId int64, keyword iat.Keyword) (map[string]string, bool, string) {
	res := true
	for _, api := range keyword.Apis {
		parameter, res, message := runApi(parameter, parameterId, api)
		if res == false {
			return parameter, res, message
		}
	}
	return parameter, res, ""
}
func runApi(parameter map[string]string, parameterId int64, api iat.Api) (map[string]string, bool, string) {
	client, err := iat.GetClient()
	if err != nil {
		fmt.Println(err)
	}
	u := iat.GetUrl(parameter, api)
	requestBody := iat.GetBody(parameter, api.Body)
	requestHeaders := iat.GetHeader(parameter, api.Headers)
	req := iat.GetRequest(u, api.Method, requestHeaders, requestBody)
	startTime := iat.GetTimestamp()
	response, err := client.Do(req)
	endTime := iat.GetTimestamp()
	if err!=nil{
		return parameter,false,err.Error()
	}
	defer response.Body.Close()
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return parameter,false,err.Error()
	}
	extractors := iat.GetExtractor(api.Extractors)
	parameter, extractors = extractor(parameter, string(responseBody), extractors)
	asserts := iat.GetAssert(api.Asserts)
	res, asserts, message := assert(parameter, response.StatusCode, string(responseBody), response.Header, asserts)
	apiResult := iat.GetApiResult(api, parameterId, requestHeaders, string(responseBody), response.Header, string(responseBody), extractors, asserts, startTime, endTime, res, message)
	iat.UploadApiResult(apiResult)
	return parameter, res, message
}

func extractor(parameter map[string]string, body string, es []iat.Extractor) (map[string]string, []iat.Extractor) {
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

func assert(parameter map[string]string, httpcode int, body string, header http.Header, asserts []iat.Assert) (bool, []iat.Assert, string) {
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
	e := iat.ParameterReplace(parameter, expect)
	if method == "CONTAINS" {
		if strings.Contains(actual, iat.ParameterReplace(parameter, expect)) {
			return true, ""
		} else {
			return false, actual + " no contains " + e
		}
	} else if method == "EQUALS" {
		if actual == iat.ParameterReplace(parameter, expect) {
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
