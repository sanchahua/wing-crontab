package youliao

type sdkPushStruct struct {
	Type string `json:"type"`
	Platform string `json:"platform"`
	Sid []int64 `json:"sid"`
	Async int `json:"async"`
	TTL int64 `json:"ttl"`
	Msg  sdkPushDataMsg `json:"msg"`
	Delay    string `json:"delay"`
	ForcePassthrough int    `json:"force_passthrough"`
}


type sdkPushAndroidExtras struct{
	MessageID string `json:"messageId"`
	DisplayType int `json:"display_type"`
	ResType string `json:"res_type"`
	Title string `json:"title"`
	Desc string `json:"desc"`
	LiveInfo interface{} `json:"live_info"`
	NotifyId int64 `json:"notify_id"`
}

type sdkPushDataMsg struct {
	Title string `json:"title"`
	Body string `json:"body"`
	Extras sdkPushAndroidExtras `json:"extras"`
}