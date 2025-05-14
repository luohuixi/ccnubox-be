package metrics

type MetricsReq struct {
	Level string `json:"level"` //错误等级,分为info,error,warn,debug四个等级
	Msg   string `json:"msg"`   //错误信息
}
