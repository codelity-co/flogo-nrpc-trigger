package nrpc

import (
	"github.com/project-flogo/core/data/coerce"
)

// Settings struct
type Settings struct {
	NatsClusterUrls          string `md:"natsClusterUrls,required"`
	NatsConnName             string `md:"natsConnName"`
	NatsUserName             string `md:"natsUserName"`
	NatsUserPassword         string `md:"natsUserPassword"`
	NatsToken                string `md:"natsToken"`
	NatsNkeySeedfile         string `md:"natsNkeySeedfile"`
	NatsCredentialFile       string `md:"natsCredentialFile"`
	AutoReconnect            bool   `md:"autoReconnect"`
	MaxReconnects            int    `md:"maxReconnects"`
	EnableRandomReconnection bool   `md:"enableRandomReconnection"`
	ReconnectWait            int    `md:"reconnectWait"`
	ReconnectBufferSize      int    `md:"reconnectBufferSize"`
	SkipVerify               bool   `md:"skipVerify"`
	CaFile                   string `md:"caFile"`
	CertFile                 string `md:"certFile"`
	KeyFile                  string `md:"keyFile"`
	EnableStreaming          bool   `md:"enableStreaming"`
	StanClusterID            string `md:"stanClusterID"`
	ProtoName                string `md:"protoName"`
	ProtoFile                stirng `md:"protoFile"`
}

// FromMap method of Settings
func (s *Settings) FromMap(values map[string]interface{}) error {

	var (
		err error
	)
	s.NatsClusterUrls, err = coerce.ToString(values["natsClusterUrls"])
	if err != nil {
		return err
	}

	s.NatsConnName, err = coerce.ToString(values["natsConnName"])
	if err != nil {
		return err
	}

	s.NatsUserName, err = coerce.ToString(values["natsUserName"])
	if err != nil {
		return err
	}

	s.NatsUserPassword, err = coerce.ToString(values["natsUserPassword"])
	if err != nil {
		return err
	}

	s.NatsToken, err = coerce.ToString(values["natsToken"])
	if err != nil {
		return err
	}

	s.NatsNkeySeedfile, err = coerce.ToString(values["natsNkeySeedfile"])
	if err != nil {
		return err
	}

	s.NatsCredentialFile, err = coerce.ToString(values["natsCredentialFile"])
	if err != nil {
		return err
	}

	s.AutoReconnect, err = coerce.ToBool(values["autoReconnect"])
	if err != nil {
		return err
	}

	s.MaxReconnects, err = coerce.ToInt(values["maxReconnects"])
	if err != nil {
		return err
	}

	s.EnableRandomReconnection, err = coerce.ToBool(values["enableRandomReconnection"])
	if err != nil {
		return err
	}

	s.ReconnectWait, err = coerce.ToInt(values["reconnectWait"])
	if err != nil {
		return err
	}

	s.ReconnectBufferSize, err = coerce.ToInt(values["reconnectBufferSize"])
	if err != nil {
		return err
	}

	s.SkipVerify, err = coerce.ToBool(values["skipVerify"])
	if err != nil {
		return err
	}

	s.CaFile, err = coerce.ToString(values["caFile"])
	if err != nil {
		return err
	}

	s.CertFile, err = coerce.ToString(values["certFile"])
	if err != nil {
		return err
	}

	s.KeyFile, err = coerce.ToString(values["keyFile"])
	if err != nil {
		return err
	}

	s.EnableStreaming, err = coerce.ToBool(values["enableStreaming"])
	if err != nil {
		return err
	}

	s.StanClusterID, err = coerce.ToString(values["stanClusterID"])
	if err != nil {
		return err
	}

	s.ProtoName, err = coerce.ToString(values["protoName"])
	if err != nil {
		return err
	}

	s.ProtoFile, err = coerce.ToString(values["protoFile"])
	if err != nil {
		return err
	}
	return nil

}

// ToMap method of Settings
func (s *Settings) ToMap() map[string]interface{} {

	return map[string]interface{}{
		"natsClusterUrls":          s.NatsClusterUrls,
		"natsConnName":             s.NatsConnName,
		"natsUserName":             s.NatsUserName,
		"natsUserPassword":         s.NatsUserPassword,
		"natsToken":                s.NatsToken,
		"natsNkeySeedfile":         s.NatsNkeySeedfile,
		"natsCredentialFile":       s.NatsNkeySeedfile,
		"autoReconnect":            s.AutoReconnect,
		"maxReconnects":            s.MaxReconnects,
		"enableRandomReconnection": s.EnableRandomReconnection,
		"reconnectWait":            s.ReconnectWait,
		"reconnectBufferSize":      s.ReconnectBufferSize,
		"skipVerify":               s.SkipVerify,
		"caFile":                   s.CaFile,
		"certFile":                 s.CertFile,
		"keyFile":                  s.KeyFile,
		"enableStreaming":          s.EnableStreaming,
		"stanClusterID":            s.StanClusterID,
		"protoName":                s.ProtoName,
		"protoFile":                s.ProtoFile,
	}

}

type Output struct {
	Data map[string]interface{} `md:"data"`
}

func (o *Output) FromMap(values map[string]interface{}) error {
	var err error

	o.Data, err = coerce.ToObject(values["data"])
	if err != nil {
		return err
	}

	return nil
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"data": o.Data,
	}
}

type Reply struct {
	Code int         `md:"code"`
	Data interface{} `md:"data"`
}

func (r *Reply) FromMap(values map[string]interface{}) error {
	var err error

	r.Code, err = coerce.ToInt(values["code"])
	if err != nil {
		return err
	}

	r.Data, err = coerce.ToObject(values["data"])
	if err != nil {
		return err
	}

	return nil
}

func (r *Reply) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"code": r.Code,
		"data": r.Data,
	}
}
