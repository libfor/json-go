package json

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unsafe"
)

var AppendSlices = false

type Nested struct {
	Amazing string
}

type TestType struct {
	Name      string
	Food      string
	Tags      map[string]string
	Nested    *Nested
	SomeList  []string
	EmptyList []string
}

//easyjson:json
type EasyType struct {
	Name      string
	Food      string
	Tags      map[string]string
	Nested    *Nested
	SomeList  []string
	EmptyList []string
}

type jsonDecoder interface {
	DecodeFrom(reflect.Value, []byte, int) error
}

type jsonStoredProcedure interface {
	IntoPointer(decodeOperation, int, int, uintptr) (int, error)
}

type describer interface {
	Describe(t reflect.Type) jsonStoredProcedure
}

var ErrIncompleteRead = errors.New("incomplete read")

var ErrNotPointer = errors.New("cannot unmarshal to a nonpointer")

var ErrUnexpectedEOF = errors.New("unexpected EOF")

var ErrNotImplemented = errors.New("not implemented")

var ErrUnexpectedListEnd = errors.New("unexpected ]")

var zeroString = reflect.ValueOf("")

type jsonRawString struct{}

func (j jsonRawString) IntoPointer(op decodeOperation, p, end int, base uintptr) (int, error) {
	b := op.rawData
	if Verbose {
		fmt.Println("looking for raw string in:", string(b[p:end]))
	}
	for p < end {
		thisChar := b[p]
		if thisChar == ']' {
			return p, ErrUnexpectedListEnd
		}
		p += 1
		if thisChar == '"' {
			start := p
			for p < end {
				thisChar := b[p]
				p += 1
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

type jsonEscapedString struct{}

func (j jsonEscapedString) IntoPointer(op decodeOperation, p, end int, base uintptr) (int, error) {
	b := op.rawData
	if Verbose {
		fmt.Println("looking for escaped string in:", string(b[p:end]))
	}
	for p < end {
		thisChar := b[p]
		if thisChar == ']' {
			if Verbose {
				fmt.Println("escaped string saw unexpected list end")
			}
			return p, ErrUnexpectedListEnd
		}
		p += 1
		if thisChar == '"' {
			start := p
			for p < end {
				thisChar := b[p]
				p += 1
				if thisChar == '"' {
					if base != 0 {
						*(*string)(unsafe.Pointer(base)) = string(b[start : p-1])
						if Verbose {
							fmt.Println("found escaped string in:", string(b[start:p-1]))
						}
					}
					return p, nil
				}
			}
		}
	}
	return end, ErrUnexpectedEOF
}

type jsonArray struct {
	sliceType    reflect.Type
	internalProc jsonStoredProcedure
	internalType reflect.Type

	blankSpace []byte
}

func newJsonArray(t reflect.Type, d describer) *jsonArray {
	e := t.Elem()
	j := &jsonArray{sliceType: t, internalProc: d.Describe(e)}

	j.internalType = e
	itemSize := int(e.Size())
	j.blankSpace = make([]byte, itemSize)
	return j
}

func (j *jsonArray) IntoPointer(op decodeOperation, p, end int, base uintptr) (int, error) {
	b := op.rawData
	if Verbose {
		fmt.Println("looking for list in", string(b[p:end]))
	}

	itemSize := int(j.internalType.Size())

	l := 0
	var pointers []byte
	posInPointers := 0

	for p < end {
		thisChar := b[p]
		if thisChar == ']' {
			return p, ErrUnexpectedListEnd
		}
		p += 1
		if thisChar == '[' {
			start := p - 1
			for p < end {
			anotherItem:
				thisChar := b[p]
				p += 1

				if thisChar == ']' {
					arr := reflect.NewAt(j.sliceType, unsafe.Pointer(base))
					currSlice := reflect.Indirect(arr)
					if posInPointers > 0 {
						amt := posInPointers / itemSize
						asArrayType := reflect.ArrayOf(amt, j.internalType)
						asArray := reflect.NewAt(asArrayType, unsafe.Pointer(&pointers[0]))
						nt := reflect.Indirect(asArray).Slice(0, amt)
						if currSlice.IsNil() {
							currSlice.Set(nt)
						} else {
							if AppendSlices {
								currSlice.Set(reflect.AppendSlice(currSlice, nt))
							} else {
								currSlice.Set(nt)
							}
						}
					} else {
						if currSlice.IsNil() {
							newArr := reflect.Indirect(reflect.MakeSlice(j.sliceType, 0, 0))
							currSlice.Set(newArr)
						}
					}
					if Verbose {
						fmt.Println("found list", string(b[start:p]))
					}
					return p, nil
				}

				posInPointers += itemSize
				if Verbose {
					fmt.Println("current pos", posInPointers, l)
				}

				if l == 0 {
					l = itemSize * 4
					pointers = make([]byte, l, l)
				}
				if cap(pointers) < posInPointers {
					l = l * 5 / 3
					if Verbose {
						fmt.Println("have to grow to", l)
					}
					np := make([]byte, l, l)
					copy(np, pointers)
					pointers = np

				}
				newPtr := unsafe.Pointer(&pointers[posInPointers-itemSize])

				n, err := j.internalProc.IntoPointer(op, p-1, end, uintptr(newPtr))
				if err != nil {
					if err != ErrUnexpectedListEnd {
						return n, err
					}
					posInPointers -= itemSize
				}
				p = n
				goto anotherItem
			}
		}
	}
	return end, ErrUnexpectedEOF
}

type jsonMaybeNull struct {
	parentType        reflect.Type
	ptrType           reflect.Type
	underlyingType    reflect.Type
	underlyingHandler jsonStoredProcedure
}

func newMaybeNull(t reflect.Type, d describer) *jsonMaybeNull {
	underT := t.Elem()
	return &jsonMaybeNull{
		parentType:        reflect.PtrTo(t),
		ptrType:           t,
		underlyingType:    underT,
		underlyingHandler: d.Describe(underT),
	}
}

func (j jsonMaybeNull) IntoPointer(op decodeOperation, p, end int, base uintptr) (int, error) {
	b := op.rawData
	// base is a pointer to a potentially-nil pointer to a T
	// we need to ensure it points to an initialized pointer to an initialized T

	if Verbose {
		fmt.Println(j.ptrType.String(), "looking for maybe null: ", string(b[p:end]))
	}

	curPtr := reflect.NewAt(j.ptrType, unsafe.Pointer(base))
	if reflect.Indirect(curPtr).IsNil() {

		newInstance := reflect.New(j.underlyingType)

		if Verbose {
			fmt.Println("found nil, curptr is", curPtr.String())
			fmt.Println("new instance", newInstance.String())
			fmt.Println("setting curptr to", newInstance.String())
			fmt.Printf("... AKA %#v\n", newInstance.Interface())
		}

		reflect.Indirect(curPtr).Set(newInstance)
	} else {
		if Verbose {
			fmt.Println("not nil", curPtr.String())
		}
	}

	n, err := j.underlyingHandler.IntoPointer(op, p, end, reflect.Indirect(curPtr).Pointer())
	if err != nil {
		return n, err
	}
	if Verbose {
		fmt.Println("found maybe null: ", string(b[p:n]))
	}
	return n, nil
}

type jsonAnything struct{}

func (j jsonAnything) IntoPointer(op decodeOperation, p, end int, base uintptr) (int, error) {
	b := op.rawData
	for p < end {
		thisChar := b[p]
		p += 1
		if thisChar == '{' {
			return (&jsonObject{}).IntoPointer(op, p-1, end, base)
		}
		if thisChar == '"' {
			return jsonRawString{}.IntoPointer(op, p-1, end, base)
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
	j.left = jsonRawString{}
	j.leftType = r.Key()

	j.right = des.Describe(r.Elem())
	j.rightType = r.Elem()

	j.all = r
	return j
}

func (j jsonMap) IntoPointer(op decodeOperation, p, end int, base uintptr) (int, error) {
	b := op.rawData
	if Verbose {
		fmt.Println("looking for map in", string(b[p:end]))
	}
	currentMap := reflect.Indirect(reflect.NewAt(j.all, unsafe.Pointer(base)))
	if reflect.Indirect(currentMap).IsNil() {
		newMap := reflect.Indirect(reflect.MakeMap(j.all))
		currentMap.Set(newMap)
	}

	// itemSize := j.rightType.Size()

	lSide := reflect.New(j.leftType)
	rSide := reflect.New(j.rightType)

	lPtr := lSide.Pointer()
	rPtr := rSide.Pointer()

	for p < end {
		thisChar := b[p]
		p += 1
		if thisChar == '{' {
			mapStart := p - 1
			for p < end {
			anotherKey:
				thisChar := b[p]
				p += 1
				if thisChar == '"' {
					n, err := j.left.IntoPointer(op, p-1, end, lPtr)
					if err != nil {
						return n, err
					}
					p = n
					for p < end {
						thisChar := b[p]
						p += 1
						if thisChar == ':' {
							n, err := j.right.IntoPointer(op, p, end, rPtr)
							if err != nil {
								return n, err
							}
							p = n
							currentMap.SetMapIndex(reflect.Indirect(lSide), reflect.Indirect(rSide))
							goto anotherKey
						}
					}
				}
				if thisChar == '}' {
					if Verbose {
						fmt.Println("found map:", string(b[mapStart:p]))
					}
					return p, nil
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
	if Verbose {
		fmt.Println(obj.String(), j.fields)
	}
	return j
}

func (j jsonObject) IntoPointer(op decodeOperation, p, end int, base uintptr) (int, error) {
	b := op.rawData
	if Verbose {
		fmt.Println("looking for object in", string(b[p:end]))
	}
	for p < end {
		thisChar := b[p]
		if thisChar == ']' {
			return p, ErrUnexpectedListEnd
		}
		p += 1
		if thisChar == '{' {
			objStart := p - 1
			for p < end {
			anotherKey:
				thisChar := b[p]
				p += 1
				if thisChar == '"' {
					start := p
					for p < end {
						thisChar := b[p]
						p += 1
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

									n, err := handler.IntoPointer(op, p, end, offset)
									if err != nil {
										return n, err
									}
									p = n
									goto anotherKey
								}
							}
						}
					}
				}
				if thisChar == '}' {
					if Verbose {
						fmt.Println("found obj in:", string(b[objStart:p]))
					}
					return p, nil
				}
			}
		}
	}
	return end, ErrUnexpectedEOF
}

func quickScan(b []byte) (ids [][3]int) {
	end := len(b)
	p := 0
	for p < end {
	anotherElement:
		thisChar := b[p]
		p += 1
		if thisChar == '"' {
			start := p - 1
			for p < end {
				thisChar := b[p]
				p += 1
				if thisChar == '\\' {
					continue
				}
				if thisChar == '"' {
					strEnd := p
					for p < end {
						thisChar := b[p]
						p += 1
						if thisChar == ':' {
							ids = append(ids, [3]int{start, strEnd, 0})
							goto anotherElement
						}
						if thisChar == ']' || thisChar == '}' || thisChar == ',' {
							ids = append(ids, [3]int{start, strEnd, 1})
							goto anotherElement
						}
					}
				}
			}
		}
	}
	return
}

type foundString struct {
	pos        int
	end        int
	destinaton *string
	raw        bool
}

type decodeOperation struct {
	rawData []byte
	strings *[]foundString
}

type Describers struct {
	allTypes     sync.Map
	pendingTypes sync.Map
}

func NewDescriber() *Describers {
	return &Describers{}
}

func (d *Describers) LearnAbout(t reflect.Type) jsonStoredProcedure {
	switch t.Kind() {
	case reflect.Ptr:
		return newMaybeNull(t, d)
	case reflect.Map:
		return newJsonMap(t, d)
	case reflect.String:
		return jsonEscapedString{}
	case reflect.Struct:
		return newJsonObject(t, d)
	case reflect.Slice:
		return newJsonArray(t, d)
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
		if Verbose {
			fmt.Println("waiting for someone to complete", t.String())
		}
		found := loading.(*sync.Cond)
		found.Wait()
		return d.Describe(t)
	}

	newProc := d.LearnAbout(t)
	if Verbose {
		fmt.Printf("storing new encoder %s: %T\n", t.String(), newProc)
	}

	d.allTypes.Store(t, newProc)
	lockIt.Broadcast()
	d.pendingTypes.Delete(t)

	return newProc
}

func (d *Describers) Unmarshal(b []byte, to interface{}) error {
	t := reflect.TypeOf(to)
	desc := d.Describe(t)

	// create a pointer to whatever i've been given
	v := reflect.ValueOf(to)

	newPtr := reflect.NewAt(t, unsafe.Pointer(v.Pointer()))

	if Verbose {
		fmt.Println("unmarshalling", newPtr.String(), newPtr.Interface(), newPtr.Pointer())
		fmt.Printf("into %T\n", desc)
	}

	reflect.Indirect(newPtr).Set(v)

	op := decodeOperation{rawData: b}
	_, err := desc.IntoPointer(op, 0, len(b), newPtr.Pointer())
	if err != nil {
		return err
	}
	return nil
}

var Standard = NewDescriber()

func Unmarshal(b []byte, to interface{}) error {
	return Standard.Unmarshal(b, to)
}
