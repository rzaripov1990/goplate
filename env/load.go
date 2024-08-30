package env

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type (
	Path struct {
		Value string
	}

	Prefix struct {
		Value string
	}
)

var ErrNoRequiredVariable = errors.New("no required variable")

func New() *BaseConfig {
	c, err := Parse[BaseConfig]()
	if err != nil {
		panic(err)
	}

	return c
}

func Parse[T_configuration any](options ...any) (configuration *T_configuration, err error) {
	paths, prefix := []string(nil), ""

	for _, option := range options {
		switch typed := option.(type) {
		case Prefix:
			prefix = typed.Value
		case Path:
			paths = append(paths, typed.Value)
		}
	}

	err = godotenv.Load(paths...)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}

	configuration = new(T_configuration)

	err = parse(reflect.ValueOf(configuration), prefix, "", "")
	if err != nil {
		panic(err)
	}

	return configuration, nil
}

func parse(value reflect.Value, name, defaul_, required string) (err error) {
	kind := value.Kind()

	switch {
	case kind == reflect.Pointer:
		return parse(value.Elem(), name, defaul_, required)
	case kind == reflect.Struct:
		typ_ := value.Type()

		if name != "" && !strings.HasSuffix(name, "_") {
			name += "_"
		}

		for i := 0; i < typ_.NumField(); i += 1 {
			field := typ_.Field(i)

			err = parse(value.Field(i), (name + field.Tag.Get("name")), field.Tag.Get("default"), field.Tag.Get("required"))
			if err != nil {
				return err
			}
		}

		return nil
	}

	if name == "" {
		return nil
	}

	read := os.Getenv(name)

	if read == "" {
		read = defaul_
	}

	if read == "" {
		if required == "true" {
			return fmt.Errorf("%w: %s", ErrNoRequiredVariable, name)
		}

		return nil
	}

	var (
		d time.Duration
		i int64
		u uint64
		f float64
		c complex128
		b bool
	)

	typeDuration := reflect.TypeOf(time.Duration(0))
	typeSliceString := reflect.TypeOf(reflect.TypeOf([]string(nil)))

	switch kind {
	case reflect.String:
		value.SetString(read)
	case reflect.Int64:
		if value.Type() == typeDuration {
			d, err = time.ParseDuration(read)
			if err != nil {
				return err
			}

			value.SetInt(int64(d))
		} else {
			i, err = strconv.ParseInt(read, 0, 64)
			if err != nil {
				return err
			}

			value.SetInt(i)
		}
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int:
		i, err = strconv.ParseInt(read, 0, 64)
		if err != nil {
			return err
		}

		value.SetInt(i)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		u, err = strconv.ParseUint(read, 0, 64)
		if err != nil {
			return err
		}

		value.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err = strconv.ParseFloat(read, 64)
		if err != nil {
			return err
		}

		value.SetFloat(f)
	case reflect.Complex64, reflect.Complex128:
		c, err = strconv.ParseComplex(read, 64)
		if err != nil {
			return err
		}

		value.SetComplex(c)
	case reflect.Bool:
		b, err = strconv.ParseBool(read)
		if err != nil {
			return err
		}

		value.SetBool(b)
	case reflect.Slice:
		if value.Type() == typeSliceString {
			data := reflect.MakeSlice(typeSliceString, 0, 0)

			for _, s := range strings.Split(read, ",") {
				data = reflect.Append(data, reflect.ValueOf(s))
			}

			value.Set(data)
		} else {
			data := reflect.MakeSlice(reflect.TypeOf([]string{}), 0, 0)

			for _, s := range strings.Split(read, ",") {
				data = reflect.Append(data, reflect.ValueOf(s))
			}

			value.Set(data)
		}
	}

	return nil
}
