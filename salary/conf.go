package salary

import (
	"encoding/json"
	"fmt"
	"github.com/zond/schedrox/common"

	"github.com/zond/sybutils/utils/gae/gaecontext"

	"google.golang.org/appengine/datastore"
)

func confKeyForDomain(k *datastore.Key) string {
	return fmt.Sprintf("SalaryConfig{Domain:%v}", k)
}

func confIdForDomain(c gaecontext.HTTPContext, dom *datastore.Key) *datastore.Key {
	return datastore.NewKey(c, "SalaryConfig", "SalaryConfig", 0, dom)
}

type Config struct {
	Id                                        *datastore.Key           `json:"id" datastore:"-"`
	SalaryPeriod                              string                   `json:"salary_period"`
	SalaryBreakpoint                          int                      `json:"salary_breakpoint"`
	SalaryReportHoursMinMinutes               int                      `json:"salary_report_hours_min_minutes"`
	SalarySerializedUserProperties            []byte                   `json:"-"`
	SalaryUserProperties                      []map[string]interface{} `json:"salary_user_properties" datastore:"-"`
	SalarySerializedParticipantTypeProperties []byte                   `json:"-"`
	SalaryParticipantTypeProperties           []map[string]interface{} `json:"salary_participant_type_properties" datastore:"-"`
	SalarySerializedEventTypeProperties       []byte                   `json:"-"`
	SalaryEventTypeProperties                 []map[string]interface{} `json:"salary_event_type_properties" datastore:"-"`
	SalarySerializedEventKindProperties       []byte                   `json:"-"`
	SalaryEventKindProperties                 []map[string]interface{} `json:"salary_event_kind_properties" datastore:"-"`
	SalarySerializedLocationProperties        []byte                   `json:"-"`
	SalaryLocationProperties                  []map[string]interface{} `json:"salary_location_properties" datastore:"-"`
	SalarySerializedCode                      []byte                   `json:"-"`
	SalaryCode                                string                   `json:"salary_code" datastore:"-"`
}

func (self *Config) process(c gaecontext.HTTPContext) *Config {
	if self.SalarySerializedParticipantTypeProperties != nil {
		if err := json.Unmarshal(self.SalarySerializedParticipantTypeProperties, &self.SalaryParticipantTypeProperties); err != nil {
			panic(err)
		}
	}
	if self.SalarySerializedUserProperties != nil {
		if err := json.Unmarshal(self.SalarySerializedUserProperties, &self.SalaryUserProperties); err != nil {
			panic(err)
		}
	}
	if self.SalarySerializedEventTypeProperties != nil {
		if err := json.Unmarshal(self.SalarySerializedEventTypeProperties, &self.SalaryEventTypeProperties); err != nil {
			panic(err)
		}
	}
	if self.SalarySerializedEventKindProperties != nil {
		if err := json.Unmarshal(self.SalarySerializedEventKindProperties, &self.SalaryEventKindProperties); err != nil {
			panic(err)
		}
	}
	if self.SalarySerializedLocationProperties != nil {
		if err := json.Unmarshal(self.SalarySerializedLocationProperties, &self.SalaryLocationProperties); err != nil {
			panic(err)
		}
	}
	self.SalaryCode = string(self.SalarySerializedCode)
	return self
}

func GetConfigs(c gaecontext.HTTPContext, domains []*datastore.Key) (result []*Config) {
	keys := make([]string, len(domains))
	destPs := make([]interface{}, len(domains))
	funcs := make([]func() interface{}, len(domains))
	for index, dom := range domains {
		keys[index] = confKeyForDomain(dom)
		var conf Config
		destPs[index] = &conf
		idCopy := dom
		funcs[index] = func() interface{} {
			return findConfig(c, idCopy)
		}
	}

	common.MemoizeMulti(c, keys, destPs, funcs)

	result = make([]*Config, len(domains))

	for index, _ := range destPs {
		result[index] = destPs[index].(*Config).process(c)
	}

	return
}

func findConfig(c gaecontext.HTTPContext, dom *datastore.Key) *Config {
	var conf Config
	key := confIdForDomain(c, dom)
	err := datastore.Get(c, key, &conf)
	if err != datastore.ErrNoSuchEntity {
		common.AssertOkError(err)
	}
	conf.Id = key
	return &conf
}

func GetConfig(c gaecontext.HTTPContext, dom *datastore.Key) *Config {
	var conf Config
	common.Memoize(c, confKeyForDomain(dom), &conf, func() interface{} {
		return findConfig(c, dom)
	})
	return (&conf).process(c)
}

func (self *Config) Save(c gaecontext.HTTPContext, dom *datastore.Key) *Config {
	var err error

	if self.SalarySerializedParticipantTypeProperties, err = json.Marshal(self.SalaryParticipantTypeProperties); err != nil {
		panic(err)
	}

	if self.SalarySerializedUserProperties, err = json.Marshal(self.SalaryUserProperties); err != nil {
		panic(err)
	}

	if self.SalarySerializedEventTypeProperties, err = json.Marshal(self.SalaryEventTypeProperties); err != nil {
		panic(err)
	}

	if self.SalarySerializedEventKindProperties, err = json.Marshal(self.SalaryEventKindProperties); err != nil {
		panic(err)
	}

	if self.SalarySerializedLocationProperties, err = json.Marshal(self.SalaryLocationProperties); err != nil {
		panic(err)
	}

	self.SalarySerializedCode = []byte(self.SalaryCode)

	_, err = datastore.Put(c, confIdForDomain(c, dom), self)
	if err != nil {
		panic(err)
	}
	common.MemDel(c, confKeyForDomain(dom))
	return self
}
