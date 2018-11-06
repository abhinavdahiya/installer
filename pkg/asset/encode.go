package asset

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// Persister is the interface implemented by assets that
// can persist themselves into valid JSON.
type Persister interface {
	Files() (map[string][]byte, error)
}

func Persist(directory string, a Asset) error {
	e := newEncodeState()
	if err := e.marshal(a); err != nil {
		return errors.Wrap(err, "failed to marshal")
	}

	for file, raw := range e {
		path := filepath.Join(directory, file)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return errors.Wrap(err, "failed to create dir")
		}
		if err := ioutil.WriteFile(path, raw, 0644); err != nil {
			return errors.Wrap(err, "failed to write file")
		}
	}
	return nil
}

type encodeState map[string][]byte

type encoderOpts struct {
	marshaller string
}
type encoderFunc func(e *encodeState, v reflect.Value, o encoderOpts) error

func newEncodeState() encodeState {
	return map[string][]byte{}
}

func (e *encodeState) marshal(i interface{}) error {
	return e.reflectValue(reflect.ValueOf(i), encoderOpts{})
}

func (e *encodeState) reflectValue(v reflect.Value, o encoderOpts) error {
	return valueEncoder(v)(e, v, o)
}

func valueEncoder(v reflect.Value) encoderFunc {
	if !v.IsValid() {
		return newInvalidValueEncoder(fmt.Errorf("invalid type"))
	}
	return typeEncoder(v.Type())
}

func newInvalidValueEncoder(err error) encoderFunc {
	return func(_ *encodeState, _ reflect.Value, _ encoderOpts) error {
		return err
	}
}

var (
	persisterType = reflect.TypeOf((*Persister)(nil)).Elem()
)

func typeEncoder(t reflect.Type) encoderFunc {
	if t.Implements(persisterType) {
		return marshalerEncoder
	}
	switch t.Kind() {
	case reflect.Interface:
		return interfaceEncoder
	case reflect.Ptr:
		return newPtrEncoder(t)
	case reflect.Struct:
		return newStructEncoder(t)
	case reflect.Map:
		return newMapEncoder(t, "")
	case reflect.Slice:
		return newSliceEncoder(t, "")
	case reflect.Array:
		return newArrayEncoder(t, "")
	case reflect.String:
		return newMarshalEncoder()
	default:
		return newInvalidValueEncoder(fmt.Errorf("invalid type"))
	}
}

func marshalerEncoder(e *encodeState, v reflect.Value, _ encoderOpts) error {
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}
	m, ok := v.Interface().(Persister)
	if !ok {
		return nil
	}
	ae, err := m.Files()
	if err != nil {
		return fmt.Errorf("error getting files for asset")
	}
	for ak, av := range ae {
		(*e)[ak] = av
	}
	return nil
}

func interfaceEncoder(e *encodeState, v reflect.Value, o encoderOpts) error {
	if v.IsNil() {
		return nil
	}
	return e.reflectValue(v.Elem(), o)
}

type ptrEncoder struct {
	elemEnc encoderFunc
}

func (pe ptrEncoder) encode(e *encodeState, v reflect.Value, o encoderOpts) error {
	if v.IsNil() {
		return nil
	}
	return pe.elemEnc(e, v.Elem(), o)
}

func newPtrEncoder(t reflect.Type) encoderFunc {
	enc := ptrEncoder{typeEncoder(t.Elem())}
	return enc.encode
}

type structEncoder struct {
	fields []field
}

func (se structEncoder) encode(e *encodeState, v reflect.Value, o encoderOpts) error {
FieldLoop:
	for i := range se.fields {
		f := &se.fields[i]

		// Find the nested struct field by following f.index.
		fv := v
		for _, i := range f.index {
			if fv.Kind() == reflect.Ptr {
				if fv.IsNil() {
					continue FieldLoop
				}
				fv = fv.Elem()
			}
			fv = fv.Field(i)
		}
		o.marshaller = f.marshaller
		f.encoder(e, fv, o)
	}
	return nil
}

func newStructEncoder(t reflect.Type) encoderFunc {
	se := structEncoder{fields: typeFields(t)}
	return se.encode
}

type field struct {
	name string

	tag   bool
	index []int
	typ   reflect.Type

	marshaller string
	encoder    encoderFunc
}

// byIndex sorts field by index sequence.
type byIndex []field

func (x byIndex) Len() int      { return len(x) }
func (x byIndex) Swap(i, j int) { x[i], x[j] = x[j], x[i] }
func (x byIndex) Less(i, j int) bool {
	for k, xik := range x[i].index {
		if k >= len(x[j].index) {
			return false
		}
		if xik != x[j].index[k] {
			return xik < x[j].index[k]
		}
	}
	return len(x[i].index) < len(x[j].index)
}

func typeFields(t reflect.Type) []field {
	current := []field{}
	next := []field{{typ: t}}

	count := map[reflect.Type]int{}
	nextCount := map[reflect.Type]int{}

	visited := map[reflect.Type]bool{}

	var fields []field

	for len(next) > 0 {
		current, next = next, current[:0]
		count, nextCount = nextCount, map[reflect.Type]int{}

		for _, f := range current {
			if visited[f.typ] {
				continue
			}
			visited[f.typ] = true

			for i := 0; i < f.typ.NumField(); i++ {
				sf := f.typ.Field(i)

				tag := sf.Tag.Get("asset")
				if tag == "-" {
					continue
				}
				name, marshaller := parseTag(tag)

				index := make([]int, len(f.index)+1)
				copy(index, f.index)
				index[len(f.index)] = i

				ft := sf.Type
				if ft.Name() == "" && ft.Kind() == reflect.Ptr {
					// Follow pointer.
					ft = ft.Elem()
				}

				if name != "" || !sf.Anonymous || ft.Kind() != reflect.Struct {
					tagged := name != ""
					if name == "" {
						name = sf.Name
					}
					field := field{
						name:       name,
						tag:        tagged,
						index:      index,
						typ:        ft,
						marshaller: marshaller,
					}
					fields = append(fields, field)
					if count[f.typ] > 1 {
						// If there were multiple instances, add a second,
						// so that the annihilation code will see a duplicate.
						// It only cares about the distinction between 1 or 2,
						// so don't bother generating any more copies.
						fields = append(fields, fields[len(fields)-1])
					}
					continue
				}

				// Record new anonymous struct to explore in next round.
				nextCount[ft]++
				if nextCount[ft] == 1 {
					next = append(next, field{name: ft.Name(), index: index, typ: ft})
				}
			}
		}
	}

	sort.Slice(fields, func(i, j int) bool {
		x := fields
		// sort field by name, breaking ties with depth, then
		// breaking ties with index sequence.
		if x[i].name != x[j].name {
			return x[i].name < x[j].name
		}
		if len(x[i].index) != len(x[j].index) {
			return len(x[i].index) < len(x[j].index)
		}
		if x[i].tag != x[j].tag {
			return x[i].tag
		}
		return byIndex(x).Less(i, j)
	})

	// Delete all fields that are hidden by the Go rules for embedded fields,
	// except that fields with tags are promoted.

	// The fields are sorted in primary order of name, secondary order
	// of field index length. Loop over names; for each name, delete
	// hidden fields by choosing the one dominant field that survives.
	out := fields[:0]
	for advance, i := 0, 0; i < len(fields); i += advance {
		// One iteration per name.
		// Find the sequence of fields with the name of this first field.
		fi := fields[i]
		name := fi.name
		for advance = 1; i+advance < len(fields); advance++ {
			fj := fields[i+advance]
			if fj.name != name {
				break
			}
		}
		if advance == 1 { // Only one field with this name
			out = append(out, fi)
			continue
		}
		dominant, ok := dominantField(fields[i : i+advance])
		if ok {
			out = append(out, dominant)
		}
	}

	fields = out
	sort.Sort(byIndex(fields))

	for i := range fields {
		f := &fields[i]
		f.encoder = fieldTypeEncoder(typeByIndex(t, f.index), f.name)
	}
	return fields
}

func typeByIndex(t reflect.Type, index []int) reflect.Type {
	for _, i := range index {
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		t = t.Field(i).Type
	}
	return t
}

func fieldTypeEncoder(t reflect.Type, name string) encoderFunc {
	switch t.Kind() {
	case reflect.Map:
		return newMapEncoder(t, name)
	case reflect.Slice:
		return newSliceEncoder(t, name)
	case reflect.Array:
		return newArrayEncoder(t, name)
	default:
		me := marshalEncoder{filename: name}
		return me.encode
	}
}

// dominantField looks through the fields, all of which are known to
// have the same name, to find the single field that dominates the
// others using Go's embedding rules, modified by the presence of
// tags. If there are multiple top-level fields, the boolean
// will be false: This condition is an error in Go and we skip all
// the fields.
func dominantField(fields []field) (field, bool) {
	// The fields are sorted in increasing index-length order, then by presence of tag.
	// That means that the first field is the dominant one. We need only check
	// for error cases: two fields at top level, either both tagged or neither tagged.
	if len(fields) > 1 && len(fields[0].index) == len(fields[1].index) && fields[0].tag == fields[1].tag {
		return field{}, false
	}
	return fields[0], true
}

// <PATH><, MARSHALLER json|yaml (default: json)>
func parseTag(t string) (path, marshaller string) {
	arr := strings.SplitN(t, ",", 2)
	path = arr[0]
	marshaller = "json"
	if len(arr) == 2 && (arr[1] == "yaml" || arr[1] == "yml") {
		marshaller = "yaml"
	}
	return
}

type mapEncoder struct {
	directory string
}

func (me mapEncoder) encode(e *encodeState, v reflect.Value, o encoderOpts) error {
	keys := v.MapKeys()
	for _, kv := range keys {
		vv := v.MapIndex(kv)

		path := kv.String()
		if me.directory != "" {
			path = filepath.Join(me.directory, path)
		}
		me := marshalEncoder{filename: path}
		if err := me.encode(e, vv, o); err != nil {
			return errors.Wrapf(err, "failed to encode key %v", kv)
		}
	}
	return nil
}

func newMapEncoder(t reflect.Type, dir string) encoderFunc {
	if t.Key().Kind() != reflect.String {
		return newInvalidValueEncoder(fmt.Errorf("only string keys are supported for maps"))
	}
	me := mapEncoder{directory: dir}
	return me.encode
}

type sliceEncoder struct {
	arrayEnc encoderFunc
}

func (se sliceEncoder) encode(e *encodeState, v reflect.Value, o encoderOpts) error {
	if v.IsNil() {
		return nil
	}
	return se.arrayEnc(e, v, o)
}

func newSliceEncoder(t reflect.Type, dir string) encoderFunc {
	if isByteSlice(t) {
		return newMarshalEncoder()
	}
	se := sliceEncoder{newArrayEncoder(t, dir)}
	return se.encode
}

type arrayEncoder struct {
	directory string
}

func (ae arrayEncoder) encode(e *encodeState, v reflect.Value, o encoderOpts) error {
	n := v.Len()
	fPrefix := v.Type().Name()
	for i := 0; i < n; i++ {
		path := fmt.Sprintf("%s-%d", fPrefix, i)
		if ae.directory != "" {
			path = filepath.Join(ae.directory, path)
		}
		me := marshalEncoder{filename: path}
		if err := me.encode(e, v.Index(i), o); err != nil {
			return errors.Wrapf(err, "failed to encode index %d", i)
		}
	}
	return nil
}

func newArrayEncoder(t reflect.Type, dir string) encoderFunc {
	ae := arrayEncoder{directory: dir}
	return ae.encode
}

type marshalEncoder struct {
	filename string
}

func (me marshalEncoder) encode(e *encodeState, v reflect.Value, o encoderOpts) error {
	if isByteSlice(v.Type()) {
		if v.IsNil() {
			return nil
		}
		if me.filename == "" {
			return fmt.Errorf("filename cannot be empty")
		}
		(*e)[me.filename] = v.Bytes()
		return nil
	}
	filename := v.Type().Name()
	if me.filename != "" {
		filename = me.filename
	}

	var out []byte
	var err error
	marshaller := filepath.Ext(filename)
	if o.marshaller != "" {
		marshaller = o.marshaller
	}
	switch marshaller {
	case "yaml", "yml":
		out, err = yaml.Marshal(v.Interface())
	default:
		out, err = json.Marshal(v.Interface())
	}
	if err != nil {
		return errors.Wrap(err, "failed to marshal")
	}
	(*e)[filename] = out
	return nil
}

func newMarshalEncoder() encoderFunc {
	enc := marshalEncoder{}
	return enc.encode
}

func isByteSlice(t reflect.Type) bool {
	if t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
		return true
	}
	return false
}
