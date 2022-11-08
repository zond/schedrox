package common

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang/snappy"

	"github.com/vmihailenco/msgpack"

	"github.com/zond/sybutils/utils/gae/memcache"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	amemcache "google.golang.org/appengine/memcache"
)

const (
	regular = iota
	nilCache
	JSTimeFormat  = "2006-01-02T15:04:05Z"
	ISO8601Format = "20060102"
)

var UTC *time.Location

var prefPattern = regexp.MustCompile("^([^\\s;]+)(;q=([\\d.]+))?$")

func Overlaps(r1, r2 [2]time.Time) bool {
	return (!r1[0].After(r2[0]) && !r1[1].Before(r2[1])) || (!r1[0].Before(r2[0]) && r1[0].Before(r2[1])) || (r1[1].After(r2[0]) && !r1[1].After(r2[1]))
}

func init() {
	var err error
	UTC, err = time.LoadLocation("UTC")
	if err != nil {
		panic(err)
	}
	rand.Seed(time.Now().UnixNano())
	memcache.Codec = snappyMsgPack
}

func EncodeKey(k *datastore.Key) string {
	return k.Encode()
}

func DecodeKey(s string) *datastore.Key {
	k, _ := datastore.DecodeKey(s)
	return k
}

var snappyMsgPack = amemcache.Codec{
	Marshal: func(i interface{}) (b []byte, err error) {
		encoded, err := msgpack.Marshal(i)
		if err != nil {
			return
		}
		if len(encoded) > 999750 {
			var compacted []byte
			if compacted, err = snappy.Encode(nil, encoded); err != nil {
				return
			}
			b = make([]byte, len(compacted)+1)
			b[0] = 1
			copy(b[1:], compacted)
		} else {
			b = make([]byte, len(encoded)+1)
			b[0] = 0
			copy(b[1:], encoded)
		}
		if len(b) > 999750 {
			err = fmt.Errorf("TOO BIG MEMCACHE BLOB: %v bytes", len(b))
		}
		return
	},
	Unmarshal: func(b []byte, i interface{}) (err error) {
		if b[0] == 0 {
			if err = msgpack.Unmarshal(b[1:], i); err != nil {
				return
			}
			return
		} else {
			var extracted []byte
			if extracted, err = snappy.Decode(nil, b[1:]); err != nil {
				return
			}
			if err = msgpack.Unmarshal(extracted, i); err != nil {
				return
			}
			return
		}
	},
}

type ByteString []byte

func (self ByteString) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(self))
}

func (self *ByteString) UnmarshalJSON(b []byte) (err error) {
	s := ""
	if err = json.Unmarshal(b, &s); err != nil {
		return
	}
	*self = []byte(s)
	return
}

func SenderUser(c context.Context) string {
	switch appengine.AppID(c) {
	case "kc-sched":
		return "klättercentret noreply"
	}
	return "schedrox noreply"
}

func SetContentType(w http.ResponseWriter, t string, cache bool) {
	w.Header().Set("Content-Type", t)
	w.Header().Set("Vary", "Accept")
	if cache {
		if !appengine.IsDevAppServer() {
			w.Header().Set("Cache-Control", "public, max-age=864000")
		}
	} else {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
	}
}

func EncodeBase64(s string) string {
	buf := new(bytes.Buffer)
	enc := base64.NewEncoder(base64.StdEncoding, buf)
	if n, err := enc.Write([]byte(s)); err != nil {
		panic(err)
	} else if n != len([]byte(s)) {
		panic(fmt.Errorf("Wanted to write %v bytes, but wrote %v bytes", len([]byte(s)), n))
	}
	enc.Close()
	return string(buf.Bytes())
}

func MustDecodeBase64(s string) string {
	dec := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(s))
	buf := new(bytes.Buffer)
	io.Copy(buf, dec)
	return string(buf.Bytes())
}

func MustParseISOTime(s string) (result time.Time) {
	var err error
	if result, err = time.Parse(ISO8601Format, s); err != nil {
		panic(err)
	}
	return
}

func MustParseJSTime(s string) (result time.Time) {
	var err error
	if result, err = time.Parse(JSTimeFormat, s); err != nil {
		panic(err)
	}
	return
}

func EncKey(k *datastore.Key) string {
	if k == nil {
		return "nil"
	}
	return k.Encode()
}

type Page struct {
	Results interface{} `json:"results"`
	Total   int         `json:"total"`
}

func Min(i ...int) (result int) {
	result = i[0]
	for _, x := range i {
		if x < result {
			result = x
		}
	}
	return
}

func Max(i ...int) (result int) {
	result = i[0]
	for _, x := range i {
		if x > result {
			result = x
		}
	}
	return
}

type Week int64

func (self Week) Year() int {
	return int(self >> 16)
}
func (self Week) Week() int {
	return int(self & ((2 << 16) - 1))
}
func NewWeek(t time.Time) Week {
	y, w := t.ISOWeek()
	return Week((int64(y) << 16) + int64(w))
}

func IsOkError(err error, accepted ...error) bool {
	acceptedMap := map[string]bool{}
	for _, e := range accepted {
		acceptedMap[e.Error()] = true
	}
	if err != nil {
		if merr, ok := err.(appengine.MultiError); ok {
			for _, serr := range merr {
				if serr != nil {
					if _, ok := serr.(*datastore.ErrFieldMismatch); !ok && !acceptedMap[serr.Error()] {
						return false
					}
				}
			}
		} else if _, ok := err.(*datastore.ErrFieldMismatch); !ok && !acceptedMap[err.Error()] {
			return false
		}
	}
	return true
}

func AssertOkError(err error, accepted ...error) {
	if !IsOkError(err, accepted...) {
		panic(err)
	}
}

func MostAccepted(r *http.Request, def, name string) string {
	most := MostAcceptedMap(r, def, name)
	if len(most) == 1 {
		for k, _ := range most {
			return k
		}
	}
	return def
}

func MostAcceptedMap(r *http.Request, def, name string) (result map[string]bool) {
	result = map[string]bool{}
	var bestScore float64 = -1
	var score float64
	for _, pref := range strings.Split(r.Header.Get(name), ",") {
		if match := prefPattern.FindStringSubmatch(pref); match != nil {
			score = 1
			if match[3] != "" {
				score = MustParseFloat64(match[3])
			}
			if score > bestScore {
				result = map[string]bool{}
				result[match[1]] = true
				bestScore = score
			} else if score == bestScore {
				result[match[1]] = true
			}
		}
	}
	if len(result) == 0 {
		result[def] = true
	}
	return
}

func isNil(v reflect.Value) bool {
	k := v.Kind()
	if k == reflect.Chan {
		return v.IsNil()
	}
	if k == reflect.Func {
		return v.IsNil()
	}
	if k == reflect.Interface {
		return v.IsNil()
	}
	if k == reflect.Map {
		return v.IsNil()
	}
	if k == reflect.Ptr {
		return v.IsNil()
	}
	if k == reflect.Slice {
		return v.IsNil()
	}
	return false
}

func keyify(k string) string {
	buf := new(bytes.Buffer)
	enc := base64.NewEncoder(base64.StdEncoding, buf)
	h := sha1.New()
	io.WriteString(h, k)
	sum := h.Sum(nil)
	if wrote, err := enc.Write(sum); err != nil {
		panic(err)
	} else if wrote != len(sum) {
		panic(fmt.Errorf("Tried to write %v bytes but wrote %v bytes", len(sum), wrote))
	}
	if err := enc.Close(); err != nil {
		panic(err)
	}
	return string(buf.Bytes())
}

func MemDel(c memcache.TransactionContext, keys ...string) {
	if err := memcache.Del(c, keys...); err != nil {
		panic(err)
	}
}

func Memoize2(c memcache.TransactionContext, super, key string, destP interface{}, f func() interface{}) (existed bool) {
	if err := memcache.Memoize2(c, super, key, destP, func() (interface{}, error) {
		return f(), nil
	}); err == nil {
		existed = true
	} else if err == memcache.ErrCacheMiss {
		existed = false
	} else {
		panic(err)
	}
	return
}

func reflectCopy(srcValue reflect.Value, source, destP interface{}) {
	if reflect.PtrTo(reflect.TypeOf(source)) == reflect.TypeOf(destP) {
		reflect.ValueOf(destP).Elem().Set(srcValue)
	} else {
		reflect.ValueOf(destP).Elem().Set(reflect.Indirect(srcValue))
	}
}

func Memoize(c memcache.TransactionContext, key string, destP interface{}, f func() interface{}) (existed bool) {
	return MemoizeMulti(c, []string{key}, []interface{}{destP}, []func() interface{}{f})[0]
}

func MustMarshalJSON(i interface{}) (result []byte) {
	var err error
	if result, err = json.Marshal(i); err != nil {
		panic(err)
	}
	return
}

func MustUnmarshalJSON(b []byte, result interface{}) {
	if err := json.Unmarshal(b, result); err != nil {
		panic(err)
	}
	return
}

func MemoizeMulti(c memcache.TransactionContext, keys []string, destPs []interface{}, f []func() interface{}) (exists []bool) {
	subverted := make([]func() (interface{}, error), len(f))
	for index, fu := range f {
		fucpy := fu
		subverted[index] = func() (interface{}, error) {
			return fucpy(), nil
		}
	}
	errs := memcache.MemoizeMulti(c, keys, destPs, subverted)
	exists = make([]bool, len(keys))
	for index, _ := range exists {
		if len(errs) > index && errs[index] != nil {
			if errs[index] == memcache.ErrCacheMiss {
				exists[index] = false
			} else {
				panic(errs)
			}
		} else {
			exists[index] = true
		}
	}
	return
}

func WeeksBetween(start, end time.Time) (result []Week) {
	if end.Before(start) {
		return nil
	}
	collector := make(map[Week]bool)
	for start.Add(-24 * time.Hour).Before(end) {
		collector[NewWeek(start)] = true
		start = start.Add(time.Hour * 24)
	}
	for week, _ := range collector {
		result = append(result, week)
	}
	return
}

func GetHostURL(r *http.Request) string {
	return fmt.Sprintf("https://%v", r.Host)
}

func JSONEqual(a, b interface{}) bool {
	abuf := new(bytes.Buffer)
	bbuf := new(bytes.Buffer)
	MustEncodeJSON(abuf, a)
	MustEncodeJSON(bbuf, b)
	return bytes.Equal(abuf.Bytes(), bbuf.Bytes())
}

func MustEncodeJSON(w io.Writer, i interface{}) {
	if err := json.NewEncoder(w).Encode(i); err != nil {
		panic(err)
	}
}

func MustDecodeJSON(r io.Reader, result interface{}) {
	if err := json.NewDecoder(r).Decode(result); err != nil {
		panic(err)
	}
}

func MustParseFloat64(s string) (result float64) {
	var err error
	if result, err = strconv.ParseFloat(s, 64); err != nil {
		panic(err)
	}
	return
}

func MustParseInt64(s string) (result int64) {
	var err error
	if result, err = strconv.ParseInt(s, 10, 64); err != nil {
		panic(err)
	}
	return
}

func MustParseInt(s string) (result int) {
	var err error
	if result, err = strconv.Atoi(s); err != nil {
		panic(err)
	}
	return
}
