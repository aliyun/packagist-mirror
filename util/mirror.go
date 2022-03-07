package util

import "encoding/json"

type Mirror struct {
	providerUrl          string
	distUrl              string
	apiIterationInterval int
}

func NewMirror(providerUrl string, distUrl string, apiIterationInterval int) (mirror *Mirror) {
	return &Mirror{
		providerUrl:          providerUrl,
		distUrl:              distUrl,
		apiIterationInterval: apiIterationInterval,
	}
}

type Dist struct {
	Path string `json:"path"`
	Url  string `json:"url"`
}

func NewDist(path, url string) *Dist {
	return &Dist{Path: path, Url: url}
}

func NewDistFromJSONString(jsonString string) (dist *Dist, err error) {
	dist = new(Dist)
	err = json.Unmarshal([]byte(jsonString), dist)
	return
}

func (dist *Dist) ToJSONString() string {
	distString, _ := json.Marshal(dist)
	return string(distString)
}

type Changes struct {
	Timestamp int            `json:"timestamp"`
	Actions   []ChangeAction `json:"actions"`
}

type ChangeAction struct {
	Type    string `json:"type"`
	Package string `json:"package"`
	Time    int    `json:"time"`
}

func NewChangeAction(type_ string, packageName string, time int) *ChangeAction {
	return &ChangeAction{
		Type:    type_,
		Package: packageName,
		Time:    time,
	}
}

func NewChangeActionFromJSONString(jsonString string) (action *ChangeAction, err error) {
	action = new(ChangeAction)
	err = json.Unmarshal([]byte(jsonString), action)
	return
}

func (action *ChangeAction) ToJSONString() string {
	jsonStr, _ := json.Marshal(action)
	return string(jsonStr)
}

type Task struct {
	Key  string `json:"key"`
	Path string `json:"path"`
	Hash string `json:"hash"`
}

func NewTask(key, path, hash string) *Task {
	return &Task{
		Key:  key,
		Path: path,
		Hash: hash,
	}
}

func NewTaskFromJSONString(jsonString string) (task *Task, err error) {
	task = new(Task)
	err = json.Unmarshal([]byte(jsonString), task)
	return
}

type Providers struct {
	Providers map[string]Hashes `json:"providers"`
}

func NewProvidersFromJSONString(jsonString string) (providers *Providers, err error) {
	providers = new(Providers)
	err = json.Unmarshal([]byte(jsonString), providers)
	return
}

type PackagesV2 struct {
	Minified string                 `json:"minified"`
	Packages map[string][]PackageV2 `json:"packages"`
}

type PackageV2 struct {
	Name              string      `json:"name"`
	Description       string      `json:"description"`
	Version           string      `json:"version"`
	VersionNormalized string      `json:"version_normalized"`
	Dist              PackageDist `json:"dist"`
	// {
	// "name":"alibabacloud/sdk",
	// "description":"Alibaba Cloud SDK for PHP - Easier to Use Alibaba Cloud in your PHP project",
	// "keywords":["library","sdk","cloud","aliyun","alibaba","alibabacloud"],
	// "homepage":"https://www.alibabacloud.com/",
	// "version":"1.8.1255",
	// "version_normalized":"1.8.1255.0",
	// "license":["Apache-2.0"],
	// "authors":[
	//	{"name":"Alibaba Cloud SDK","email":"sdk-team@alibabacloud.com","homepage":"http://www.alibabacloud.com"}
	// ],
	// "source":{
	//	"url":"https://github.com/aliyun/openapi-sdk-php.git",
	//	"type":"git",
	//	"reference":"a870da24afe2239d14a101214e219271a7e893b3"
	// },
	// "dist":{
	//	"url":"https://api.github.com/repos/aliyun/openapi-sdk-php/zipball/a870da24afe2239d14a101214e219271a7e893b3",
	//	"type":"zip",
	//	"shasum":"",
	//	"reference":"a870da24afe2239d14a101214e219271a7e893b3"
	//	},
	// "type":"library",
	// "time":"2022-02-10T01:24:28+00:00",
	// "autoload":{
	//	"psr-4":{"AlibabaCloud\\":"src"}
	// },
	// "require":{"php":">=5.5","ext-curl":"*","ext-json":"*","ext-libxml":"*","ext-openssl":"*","ext-mbstring":"*","ext-xmlwriter":"*","ext-simplexml":"*","alibabacloud/client":"^1.5"},
	// "require-dev":{"symfony/dotenv":"^3.4","league/climate":"^3.2.4","phpunit/phpunit":"^4.8","composer/composer":"^1.8","symfony/var-dumper":"^3.4"},
	// "suggest":{"ext-sockets":"To use client-side monitoring"},
	// "replace":{
	//	"alibabacloud/arms":"self.version",
	//	"alibabacloud/arms4finance":"self.version",
	//	"alibabacloud/aas":"self.version",
	//	"alibabacloud/actiontrail":"self.version",
	//	"alibabacloud/adb":"self.version",
	//	"alibabacloud/aegis":"self.version","alibabacloud/afs":"self.version","alibabacloud/airec":"self.version","alibabacloud/aliprobe":"self.version","alibabacloud/alidns":"self.version","alibabacloud/alikafka":"self.version","alibabacloud/alimt":"self.version","alibabacloud/aliyuncvc":"self.version","alibabacloud/appmallsservice":"self.version","alibabacloud/baas":"self.version","alibabacloud/batchcompute":"self.version","alibabacloud/bss":"self.version","alibabacloud/bssopenapi":"self.version","alibabacloud/ccc":"self.version","alibabacloud/cf":"self.version","alibabacloud/cs":"self.version","alibabacloud/csb":"self.version","alibabacloud/cas":"self.version","alibabacloud/cbn":"self.version","alibabacloud/ccs":"self.version","alibabacloud/cdn":"self.version","alibabacloud/cds":"self.version","alibabacloud/chatbot":"self.version","alibabacloud/cloudapi":"self.version","alibabacloud/cloudphoto":"self.version",
	//	"alibabacloud/cloudauth":"self.version","alibabacloud/cloudesl":"self.version","alibabacloud/cloudmarketing":"self.version","alibabacloud/cloudwf":"self.version","alibabacloud/cms":"self.version","alibabacloud/commondriver":"self.version","alibabacloud/companyreg":"self.version","alibabacloud/cr":"self.version","alibabacloud/crm":"self.version","alibabacloud/cusanalyticsconline":"self.version","alibabacloud/dataworkspublic":"self.version","alibabacloud/dbs":"self.version","alibabacloud/dcdn":"self.version","alibabacloud/dds":"self.version","alibabacloud/democenter":"self.version","alibabacloud/dm":"self.version","alibabacloud/dmsenterprise":"self.version","alibabacloud/domain":"self.version","alibabacloud/domainintl":"self.version","alibabacloud/drcloud":"self.version","alibabacloud/drds":"self.version","alibabacloud/dts":"self.version","alibabacloud/dybaseapi":"self.version","alibabacloud/dyplsapi":"self.version","alibabacloud/dypnsapi":"self.version","alibabacloud/dysmsapi":"self.version","alibabacloud/dyvmsapi":"self.version","alibabacloud/ehpc":"self.version","alibabacloud/eci":"self.version","alibabacloud/ecs":"self.version","alibabacloud/ecsinc":"self.version","alibabacloud/edas":"self.version","alibabacloud/elasticsearch":"self.version","alibabacloud/emr":"self.version","alibabacloud/ess":"self.version","alibabacloud/facebody":"self.version","alibabacloud/fnf":"self.version","alibabacloud/foas":"self.version","alibabacloud/ft":"self.version","alibabacloud/goodstech":"self.version","alibabacloud/gpdb":"self.version","alibabacloud/green":"self.version","alibabacloud/hbase":"self.version","alibabacloud/hpc":"self.version","alibabacloud/hiknoengine":"self.version","alibabacloud/hsm":"self.version","alibabacloud/httpdns":"self.version","alibabacloud/idst":"self.version","alibabacloud/itaas":"self.version","alibabacloud/imagesearch":"self.version","alibabacloud/imageaudit":"self.version","alibabacloud/imageenhan":"self.version","alibabacloud/imagerecog":"self.version","alibabacloud/imageseg":"self.version","alibabacloud/imm":"self.version","alibabacloud/industrybrain":"self.version","alibabacloud/iot":"self.version","alibabacloud/iqa":"self.version","alibabacloud/ivision":"self.version","alibabacloud/ivpd":"self.version","alibabacloud/jaq":"self.version","alibabacloud/jarvis":"self.version","alibabacloud/jarvispublic":"self.version","alibabacloud/kms":"self.version","alibabacloud/linkface":"self.version","alibabacloud/linkwan":"self.version","alibabacloud/linkedmall":"self.version","alibabacloud/live":"self.version","alibabacloud/lubancloud":"self.version","alibabacloud/lubanruler":"self.version","alibabacloud/mpserverless":"self.version","alibabacloud/market":"self.version","alibabacloud/mopen":"self.version","alibabacloud/mts":"self.version","alibabacloud/multimediaai":"self.version","alibabacloud/nas":"self.version","alibabacloud/netana":"self.version","alibabacloud/nlp":"self.version","alibabacloud/nlpautoml":"self.version","alibabacloud/nlscloudmeta":"self.version","alibabacloud/nlsfiletrans":"self.version","alibabacloud/objectdet":"self.version","alibabacloud/ocr":"self.version","alibabacloud/ocs":"self.version","alibabacloud/oms":"self.version","alibabacloud/ons":"self.version","alibabacloud/onsmqtt":"self.version","alibabacloud/oos":"self.version","alibabacloud/openanalytics":"self.version","alibabacloud/ossadmin":"self.version","alibabacloud/ots":"self.version","alibabacloud/outboundbot":"self.version","alibabacloud/pts":"self.version","alibabacloud/petadata":"self.version","alibabacloud/polardb":"self.version","alibabacloud/productcatalog":"self.version","alibabacloud/push":"self.version","alibabacloud/pvtz":"self.version","alibabacloud/qualitycheck":"self.version","alibabacloud/rkvstore":"self.version","alibabacloud/ros":"self.version","alibabacloud/ram":"self.version","alibabacloud/rds":"self.version","alibabacloud/reid":"self.version","alibabacloud/retailcloud":"self.version","alibabacloud/rtc":"self.version","alibabacloud/saf":"self.version","alibabacloud/sas":"self.version","alibabacloud/sasapi":"self.version","alibabacloud/scdn":"self.version","alibabacloud/schedulerx2":"self.version","alibabacloud/skyeye":"self.version","alibabacloud/slb":"self.version","alibabacloud/smartag":"self.version","alibabacloud/smc":"self.version","alibabacloud/sms":"self.version","alibabacloud/smsintl":"self.version","alibabacloud/snsuapi":"self.version","alibabacloud/sts":"self.version","alibabacloud/taginner":"self.version","alibabacloud/tesladam":"self.version","alibabacloud/teslamaxcompute":"self.version","alibabacloud/teslastream":"self.version","alibabacloud/ubsms":"self.version","alibabacloud/ubsmsinner":"self.version","alibabacloud/uis":"self.version","alibabacloud/unimkt":"self.version","alibabacloud/visionai":"self.version","alibabacloud/vod":"self.version","alibabacloud/voicenavigator":"self.version","alibabacloud/vpc":"self.version","alibabacloud/vs":"self.version","alibabacloud/wafopenapi":"self.version","alibabacloud/welfareinner":"self.version","alibabacloud/xspace":"self.version","alibabacloud/xtrace":"self.version","alibabacloud/yqbridge":"self.version","alibabacloud/yundun":"self.version"},
	// "support":{
	//	"issues":"https://github.com/aliyun/openapi-sdk-php/issues",
	//	"source":"https://github.com/aliyun/openapi-sdk-php"
	// }
	// }
}
