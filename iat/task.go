package iat

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

func GetExtractor(extractor string) []Extractor {
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

func GetAssert(assert string) []Assert {
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

func GetHeader(parameter map[string]string, header string) map[string]string {
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
		hds[h.Name] = ParameterReplace(parameter, h.DefaultValue)
	}
	return hds
}

func GetBody(parameter map[string]string, body string) string {
	if body == "" {
		return ""
	}
	var b Body
	err := json.Unmarshal([]byte(body), &b)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return ParameterReplace(parameter, b.DefaultValue)
}

func GetFormData(parameter map[string]string, formData string) map[string]string {
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
		form[f.Name] = ParameterReplace(parameter, f.DefaultValue)
	}
	return form
}



func GetParameter(p map[string]string, parameter Parameter) map[string]string {
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


func ParameterReplace(parameter map[string]string, str string) string {
	re := regexp.MustCompile("#{(.*?)}+")
	sub := re.FindAllStringSubmatch(str, -1)
	for _, a := range sub {
		if parameter[a[1]] != "" {
			str = strings.Replace(str, a[0], parameter[a[1]], -1)
		}
	}
	return str
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
