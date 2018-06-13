package json

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unsafe"
)

type Nested struct {
	Amazing string
}

type TestType struct {
	Name   string
	Food   string
	Tags   map[string]string
	Nested *Nested
}

//easyjson:json
type EasyType struct {
	Name   string
	Food   string
	Tags   map[string]string
	Nested *Nested
}

type jsonDecoder interface {
	DecodeFrom(reflect.Value, []byte, int) error
}

type jsonStoredProcedure interface {
	FromZero([]byte, int, int, uintptr) (int, error)
}

type describer interface {
	Describe(t reflect.Type) jsonStoredProcedure
}

var ErrIncompleteRead = errors.New("incomplete read")

var ErrNotPointer = errors.New("cannot unmarshal to a nonpointer")

var ErrUnexpectedEOF = errors.New("unexpected EOF")

var ErrNotImplemented = errors.New("not implemented")

var zeroString = reflect.ValueOf("")

type jsonString struct{}

func (j jsonString) FromZero(b []byte, p, end int, base uintptr) (int, error) {
	for p < end {
		thisChar := b[p]
		p += 1
		if thisChar == '"' {
			start := p
			for p < end {
				thisChar := b[p]
				p += 1
				if thisChar == '\\' {
					continue
				}
				if thisChar == '"' {
					if base != 0 {
						*(*string)(unsafe.Pointer(base)) = string(b[start : p-1])
					}
					return p, nil
				}
			}
		}
	}
	return end, ErrUnexpectedEOF
}

func (j jsonString) FromPointer(b []byte, p, end int, base uintptr) (int, error) {
	return end, ErrNotImplemented
}

type jsonMaybeNull struct {
	underlyingType    reflect.Type
	underlyingHandler jsonStoredProcedure
}

func newMaybeNull(t reflect.Type, d describer) *jsonMaybeNull {
	underT := t.Elem()
	return &jsonMaybeNull{
		underlyingType:    underT,
		underlyingHandler: d.Describe(underT),
	}
}

func (j jsonMaybeNull) FromZero(b []byte, p, end int, base uintptr) (int, error) {
	curObj := reflect.NewAt(j.underlyingType, unsafe.Pointer(base))
	if curObj.IsNil() {
		newMap := reflect.Indirect(reflect.MakeMap(j.underlyingType))
		curObj.Set(newMap)
	}
	return j.underlyingHandler.FromZero(b, p, end, curObj.Pointer())
}

func (j jsonMaybeNull) FromPointer(b []byte, p, end int, base uintptr) (int, error) {
	return end, ErrNotImplemented
}

type jsonAnything struct{}

func (j jsonAnything) FromZero(b []byte, p, end int, base uintptr) (int, error) {
	for p < end {
		thisChar := b[p]
		p += 1
		if thisChar == '{' {
			return (&jsonObject{}).FromZero(b, p-1, end, base)
		}
		if thisChar == '"' {
			return jsonString{}.FromZero(b, p-1, end, base)
		}
	}
	return 0, ErrUnexpectedEOF
}

var jsonSkip = jsonAnything{}

type jsonMap struct {
	all reflect.Type

	left     jsonStoredProcedure
	leftType reflect.Type

	right     jsonStoredProcedure
	rightType reflect.Type
}

func newJsonMap(r reflect.Type, des describer) *jsonMap {
	j := &jsonMap{}
	j.left = des.Describe(r.Key())
	j.leftType = r.Key()

	j.right = des.Describe(r.Elem())
	j.rightType = r.Elem()

	j.all = r
	return j
}

func (j *jsonMap) FromZero(b []byte, p, end int, base uintptr) (int, error) {
	currentMap := reflect.Indirect(reflect.NewAt(j.all, unsafe.Pointer(base)))
	if reflect.Indirect(currentMap).IsNil() {
		newMap := reflect.Indirect(reflect.MakeMap(j.all))
		currentMap.Set(newMap)
	}
	lSide := reflect.New(j.leftType)
	rSide := reflect.New(j.rightType)
	lPtr := lSide.Pointer()
	rPtr := rSide.Pointer()

	for p < end {
		thisChar := b[p]
		p += 1
		if thisChar == '{' {
			for p < end {
			anotherKey:
				thisChar := b[p]
				p += 1
				if thisChar == '"' {
					n, err := j.left.FromZero(b, p-1, end, lPtr)
					if err != nil {
						return n, err
					}
					p = n
					for p < end {
						thisChar := b[p]
						p += 1
						if thisChar == ':' {
							n, err := j.right.FromZero(b, p, end, rPtr)
							if err != nil {
								return n, err
							}
							p = n
							currentMap.SetMapIndex(reflect.Indirect(lSide), reflect.Indirect(rSide))
							for p < end {
								thisChar := b[p]
								p += 1
								if thisChar == ',' {
									goto anotherKey
								}
								if thisChar == '}' {
									return p, nil
								}
							}
						}
					}
				}
			}
		}
	}
	return end, ErrUnexpectedEOF
}

type jsonObject struct {
	fields  map[string]uintptr
	offsets []jsonStoredProcedure
}

func newJsonObject(obj reflect.Type, des describer) *jsonObject {
	offsets := make([]jsonStoredProcedure, int(obj.Size()), int(obj.Size()))
	fields := make(map[string]uintptr, obj.NumField())
	j := &jsonObject{offsets: offsets, fields: fields}
	for i := 0; i < obj.NumField(); i++ {
		f := obj.Field(i)
		fields[strings.ToLower(f.Name)] = f.Offset
		fields[strings.ToUpper(f.Name)] = f.Offset
		fields[f.Name] = f.Offset
		offsets[f.Offset] = des.Describe(f.Type)
	}
	fmt.Println(obj.String(), j.fields)
	return j
}

func (j *jsonObject) FromZero(b []byte, p, end int, base uintptr) (int, error) {
	for p < end {
		thisChar := b[p]
		p += 1
		if thisChar == '{' {
			for p < end {
			anotherKey:
				thisChar := b[p]
				p += 1
				if thisChar == '"' {
					start := p
					for p < end {
						thisChar := b[p]
						p += 1
						if thisChar == '\\' {
							continue
						}
						if thisChar == '"' {
							key := string(b[start : p-1])
							for p < end {
								thisChar := b[p]
								p += 1
								if thisChar == ':' {
									var handler jsonStoredProcedure
									var offset uintptr
									handler = jsonSkip

									fieldOffset, ok := j.fields[key]
									if ok {
										offset = base + fieldOffset
										handler = j.offsets[fieldOffset]
									}

									n, err := handler.FromZero(b, p, end, offset)
									if err != nil {
										return n, err
									}
									p = n
									for p < end {
										thisChar := b[p]
										p += 1
										if thisChar == ',' {
											goto anotherKey
										}
										if thisChar == '}' {
											return p, nil
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return end, ErrUnexpectedEOF
}

// func (j *jsonObject) Apply(base uintptr, r io.Reader) (total int, err error) {
// 	// if n, err := r.Read(make([]byte, 10)); err != nil {
// 	// 	return total, err
// 	// } else {
// 	// 	total += n
// 	// }

// 	// b := make([]byte, 5)
// 	// if n, err := r.Read(b); err != nil {
// 	// 	return total, err
// 	// } else {
// 	// 	total += n
// 	// }

// 	// *(*string)(unsafe.Pointer(base + j.NameOffset)) = string(b)

// 	// if n, err := r.Read(make([]byte, 2)); err != nil {
// 	// 	return total, err
// 	// } else {
// 	// 	total += n
// 	// }

// 	// return
// }

type Describers struct {
	allTypes     sync.Map
	pendingTypes sync.Map
}

func NewDescriber() *Describers {
	return &Describers{}
}

func (d *Describers) LearnAbout(t reflect.Type) jsonStoredProcedure {
	fmt.Println("learning about", t)
	switch t.Kind() {
	case reflect.Ptr:
		return newMaybeNull(t, d)
	case reflect.Map:
		return newJsonMap(t, d)
	case reflect.String:
		return jsonString{}
	case reflect.Struct:
		return newJsonObject(t, d)
	default:
		panic(fmt.Sprintf("unhandled type %s", t))
	}
}

func (d *Describers) Describe(t reflect.Type) jsonStoredProcedure {
	use, found := d.allTypes.Load(t)
	if found {
		return use.(jsonStoredProcedure)
	}

	l := sync.Mutex{}
	lockIt := sync.NewCond(&l)
	l.Lock()

	loading, already := d.pendingTypes.LoadOrStore(t, lockIt)
	if already {
		fmt.Println("waiting for someone to complete", t.String())
		found := loading.(*sync.Cond)
		found.Wait()
		return d.Describe(t)
	}

	newProc := d.LearnAbout(t)
	fmt.Printf("storing new encoder %s: %T\n", t.String(), newProc)

	d.allTypes.Store(t, newProc)
	lockIt.Broadcast()
	d.pendingTypes.Delete(t)

	return newProc
}

func (d *Describers) Unmarshal(b []byte, to interface{}) error {
	desc := d.Describe(reflect.TypeOf(to))
	_, err := desc.FromZero(b, 0, len(b), reflect.ValueOf(to).Pointer())
	if err != nil {
		return err
	}
	return nil
}

var Standard = NewDescriber()

func Unmarshal(b []byte, to interface{}) error {
	return Standard.Unmarshal(b, to)
}
