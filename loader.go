package lazyflag

import (
	"encoding"
	"reflect"
	"strconv"
	"strings"
)

// found mode
const (
	Long     = 1
	Short    = 2
	Required = 4
)
const default_tag_name = "json"

type Loader struct {
	m             map[string][]string
	shortm        map[string][]string
	wordTranslate func(string) string
	tagName       string
}

func NewLoader() *Loader {
	return &Loader{
		m:             make(map[string][]string),
		shortm:        make(map[string][]string),
		tagName:       default_tag_name,
		wordTranslate: camel2hungarian,
	}
}

// if the arg value is zero, the target attribute will not be set
func (l *Loader) SetAttr(tagname string, wordTranslate func(string) string) {
	if len(tagname) == 0 {
		l.tagName = tagname
	}
	if wordTranslate != nil {
		l.wordTranslate = wordTranslate
	}
}
func NewLoaderFromStringSlice(args []string) (*Loader, error) {
	l := NewLoader()
	err := l.Parse(args)
	if err != nil {
		return nil, err
	}
	return l, nil
}
func (l *Loader) LoadAsBool(mode int, key string) bool {
	return l.iscall(mode, key)
}
func (l *Loader) LazyLoad(mode int, key string) []string {
	if mode&Long > 0 {
		if v, ok := l.m[key]; ok && len(v) > 0 {
			return v
		}
	}
	if mode&Short > 0 {
		if v, ok := l.shortm[key]; ok && len(v) > 0 {
			return v
		}
	}

	return []string{}
}
func (l *Loader) iscall(mode int, key string) bool {
	if mode&Long > 0 {
		_, ok := l.m[key]
		if ok {
			return true
		}
	}
	if mode&Short > 0 {
		_, ok := l.shortm[key]
		if ok {
			return true
		}
	}
	return false
}
func (l *Loader) LoadString(mode int, key string) (string, error) {
	v := l.LazyLoad(mode, key)
	if len(v) == 1 {
		return v[0], nil
	} else if len(v) == 0 {
		if mode&Required > 0 {
			return "", NotFound
		}
		return "", nil
	}
	return "", ErrDuplicate
}
func (l *Loader) LoadAsFloat64(mode int, key string, bitsize int) (float64, error) {
	raw, err := l.LoadString(mode, key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(raw, bitsize)
}
func (l *Loader) LoadAsInt(mode int, key string, bitsize int) (int64, error) {
	raw, err := l.LoadString(mode, key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(raw, 10, bitsize)
}
func (l *Loader) LoadAsUint(mode int, key string, bitsize int) (uint64, error) {
	raw, err := l.LoadString(mode, key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(raw, 10, bitsize)
}
func (l *Loader) LoadAs(mode int, key string, u encoding.TextUnmarshaler) error {
	v := l.LazyLoad(mode, key)
	if len(v) == 0 {
		if mode&Required > 0 {
			return NotFound
		}
		return nil
	} else if len(v) > 1 {
		return ErrDuplicate
	}
	return u.UnmarshalText([]byte(v[0]))
}
func LoadAsSlice[T encoding.TextUnmarshaler](l *Loader, mode int, key string) (result []T, err error) {
	raw := l.LazyLoad(mode, key)
	var t T
	tp := reflect.TypeOf(t).Elem()
	if len(raw) == 0 {
		if mode&Required > 0 {
			err = NotFound
		}
		return
	}
	result = make([]T, len(raw))
	for i := range raw {
		result[i] = reflect.New(tp).Interface().(T)
		err = result[i].UnmarshalText([]byte(raw[i]))
		if err != nil {
			return
		}
	}
	return
}
func (l *Loader) Parse(args []string) error {
	var (
		prefix      string
		shortprefix string
	)
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "--") {
			prefix = args[i][2:]
			if vv, ok := l.m[prefix]; !ok {
				l.m[prefix] = []string{}
			} else if len(vv) == 0 {
				return ErrDuplicate
			}
		} else if strings.HasPrefix(args[i], "-") {
			shortprefix = args[i][1:]
			if _, ok := l.shortm[shortprefix]; !ok {
				l.shortm[shortprefix] = []string{}
			}
		} else if len(prefix) > 0 {
			l.m[prefix] = append(l.m[prefix], args[i])
			prefix = ""
		} else if len(shortprefix) > 0 {
			l.shortm[prefix] = append(l.shortm[prefix], args[i])
			shortprefix = ""
		}
	}
	return nil
}

func (l *Loader) Bind(v any) error {
	return l.bind(reflect.TypeOf(v), reflect.ValueOf(v))
}

func (l *Loader) bind(vtp reflect.Type, vvl reflect.Value) error {
	// if vtp.Kind() != reflect.Pointer {
	// 	panic("bind dst must be strcut pointer")
	// }
	if vtp.Kind() == reflect.Pointer {
		vtp = vtp.Elem()
	}
	if vtp.Kind() != reflect.Struct {
		panic("bind dst must be strcut pointer")
	}
	vvl = vvl.Elem()
	fieldcount := vtp.NumField()
	for i := 0; i < fieldcount; i++ {
		field := vtp.Field(i)
		fieldv := vvl.Field(i)
		tag := field.Tag.Get(l.tagName)
		_name := l.wordTranslate(field.Name)
		if len(tag) > 0 {
			_name = tag
		}
		if fieldv.CanAddr() {
			u, ok := fieldv.Addr().Interface().(encoding.TextUnmarshaler)
			if ok {
				err := l.LoadAs(Long, _name, u)
				if err != nil {
					return err
				}
				continue
			}
		}
		switch field.Type.Kind() {
		case reflect.Struct, reflect.Pointer:
			l.bind(field.Type, fieldv)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			vval, err := l.LoadString(Long, _name)
			if err != nil {
				return err
			}
			rval, err := int_table[fieldv.Kind()](vval)
			if err != nil {
				return err
			}
			fieldv.SetInt(rval)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			vval, err := l.LoadString(Long, _name)
			if err != nil {
				return err
			}
			rval, err := uint_table[fieldv.Kind()](vval)
			if err != nil {
				return err
			}
			fieldv.SetUint(rval)
		case reflect.Bool:
			val := l.LoadAsBool(Long, _name)
			fieldv.SetBool(val)
		case reflect.String:
			val, err := l.LoadString(Long, _name)
			if err != nil {
				return err
			}
			fieldv.SetString(val)
		default:
			return TypeNotSupport + " type " + strerror(field.Type.Kind().String())
		}
	}
	return nil
}

type LoaderAttr struct {
	tagName       string
	wordTranslate func(string) string
}

func NewLoaderAttr(tagname string, wordTranslate func(string) string) *LoaderAttr {
	la := &LoaderAttr{
		tagName:       tagname,
		wordTranslate: wordTranslate,
	}
	if len(la.tagName) == 0 {
		la.tagName = "json"
	}
	if la.wordTranslate == nil {
		la.wordTranslate = camel2hungarian
	}
	return la
}

var (
	int_table = map[reflect.Kind]func(string) (int64, error){
		reflect.Int:   parseInt(32),
		reflect.Int8:  parseInt(8),
		reflect.Int16: parseInt(16),
		reflect.Int32: parseInt(32),
		reflect.Int64: parseInt(64),
	}
	uint_table = map[reflect.Kind]func(string) (uint64, error){
		reflect.Uint:   parseUint(32),
		reflect.Uint8:  parseUint(8),
		reflect.Uint16: parseUint(16),
		reflect.Uint32: parseUint(32),
		reflect.Uint64: parseUint(64),
	}
	float_table = map[reflect.Kind]func(string) (float64, error){
		reflect.Float32: parseFloat(32),
		reflect.Float64: parseFloat(64),
	}
)

func StringLoad(getter func(string) string, loader *LoaderAttr, v any) error {
	return stringload(getter, loader, reflect.ValueOf(v))
}
func stringload(getter func(string) string, loader *LoaderAttr, vvl reflect.Value) error {
	vtp := vvl.Type()
	if vtp.Kind() == reflect.Pointer {
		vtp = vtp.Elem()
		vvl = vvl.Elem()
	}
	if vtp.Kind() != reflect.Struct {
		panic("not support type " + vtp.Kind().String())
	}
	fieldcount := vtp.NumField()
	for i := 0; i < fieldcount; i++ {
		field := vtp.Field(i)
		fieldv := vvl.Field(i)
		//get name
		name := field.Tag.Get(loader.tagName)
		if len(name) == 0 {
			name = loader.wordTranslate(field.Name)
		}
		strval := getter(name)
		if len(strval) == 0 {
			continue
		}
		if fieldv.CanAddr() {
			u, ok := fieldv.Addr().Interface().(encoding.TextUnmarshaler)
			if ok {
				err := u.UnmarshalText([]byte(strval))
				if err != nil {
					return err
				}
				continue
			}
		}
		switch field.Type.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			rval, err := int_table[fieldv.Kind()](strval)
			if err != nil {
				return err
			}
			fieldv.SetInt(rval)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			rval, err := uint_table[fieldv.Kind()](strval)
			if err != nil {
				return err
			}
			fieldv.SetUint(rval)
		case reflect.Bool:
			rval, err := strconv.ParseBool(strval)
			if err != nil {
				return err
			}
			fieldv.SetBool(rval)
		case reflect.Float32, reflect.Float64:
			rval, err := float_table[fieldv.Kind()](strval)
			if err != nil {
				return err
			}
			fieldv.SetFloat(rval)
		case reflect.String:
			fieldv.SetString(strval)
		default:
			return TypeNotSupport + " " + strerror(fieldv.Kind().String())
		}
	}
	return nil
}

func parseInt(bitsize int) func(string) (int64, error) {
	return func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, bitsize)
	}
}

func parseUint(bitsize int) func(string) (uint64, error) {
	return func(s string) (uint64, error) {
		return strconv.ParseUint(s, 10, bitsize)
	}
}
func parseFloat(bitsize int) func(string) (float64, error) {
	return func(s string) (float64, error) {
		return strconv.ParseFloat(s, bitsize)
	}
}
