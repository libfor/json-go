package json

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
	"unsafe"
)

const DefaultMapSize = 4

const AppendSlices = false

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

	SurpriseMe interface{}
}

func init() {
	if Verbose {
		fmt.Printf("libfor/json[verbose:true,dryrun:%t,lookahead:%t]\n", DryRun, LookAhead)
	}
}

type jsonReport struct {
	depth    int
	messages []string
	depths   []int
}

func (j *jsonReport) Deeper() func() {
	j.depth += 1
	return func() {
		j.depth -= 1
	}
}

func (j jsonReport) String() string {
	strs := []string{}
	for i, s := range j.messages {
		strs = append(strs, strings.Repeat("  ", j.depths[i])+s)
	}
	return strings.Join(strs, "\n")
}

func (j *jsonReport) Then(format string, args ...interface{}) {
	j.messages = append(j.messages, fmt.Sprintf(format, args...))
	j.depths = append(j.depths, j.depth)
}

type jsonDecoder interface {
	DecodeFrom(reflect.Value, []byte, int) error
}

type jsonStoredProcedure interface {
	IntoPointer(decodeOperation, int, int, uintptr) (int, error)
	ReportPlan(*jsonReport)
}

type describer interface {
	ReportPlan(i interface{}) jsonReport
	Describe(t reflect.Type) jsonStoredProcedure
}

var ErrIncompleteRead = errors.New(`incomplete read`)

var ErrNotPointer = errors.New(`cannot unmarshal to a nonpointer`)

var ErrUnexpectedEOF = errors.New(`unexpected EOF`)

var ErrNotImplemented = errors.New(`not implemented`)

var ErrUnexpectedListEnd = errors.New(`unexpected ]`)

var ErrUnexpectedMapEnd = errors.New(`unexpected }`)

var ErrNoBracket = errors.New(`expected ]`)

var ErrNoBrace = errors.New(`expected }`)

var ErrNoQuote = errors.New(`expected closing "`)

var ErrNoColon = errors.New(`expected :`)

var ErrNoBracketOpen = errors.New(`expected [`)

var ErrNoBraceOpen = errors.New(`expected {`)

var ErrNoQuoteOpen = errors.New(`expected opening "`)

var zeroString = reflect.ValueOf("")

type jsonRawString struct{}

func (j jsonRawString) IntoPointer(op decodeOperation, p, end int, base uintptr) (int, error) {
	if base == 0 && op.mode != ModeSkip {
		panic("bad value here " + fmt.Sprintf("%#v", op.mode))
	}
	b := op.rawData
	if Verbose {
		if op.mode == ModeSkip {
			fmt.Println(fmt.Sprintf("%T", j), "discarding raw string in:", string(b[p:end]))
		} else {
			fmt.Println(fmt.Sprintf("%T", j), "consuming raw string in:", string(b[p:end]))
		}
	}
	for p < end {
		thisChar := b[p]
		if thisChar == ']' {
			return p, ErrUnexpectedListEnd
		}
		if thisChar == '}' {
			return p, ErrUnexpectedMapEnd
		}
		p += 1
		if thisChar == '"' {
			start := p
			for p < end {
				thisChar := b[p]
				p += 1
				if thisChar == '"' {
					if Verbose {
						fmt.Printf("found raw string in: %#v\n", string(b[start-1:p]))
					}
					if base != 0 {
						*(*string)(unsafe.Pointer(base)) = string(b[start : p-1])
					}
					return p, nil
				}
			}
			return end, ErrNoQuote
		}
	}
	return end, ErrNoQuoteOpen
}

func (j jsonRawString) ReportPlan(r *jsonReport) {
	r.Then(`Search for ", returning if I find } or ]`)
	r.Then(`Search for closing "`)
	r.Then(`Create a string in the base from the bytes I found`)
}

type jsonEscapedString struct{}

func (j jsonEscapedString) IntoPointer(op decodeOperation, p, end int, base uintptr) (int, error) {
	if base == 0 && op.mode != ModeSkip {
		panic("bad value here " + fmt.Sprintf("%#v", op.mode))
	}
	b := op.rawData
	if Verbose {
		if op.mode == ModeSkip {
			fmt.Println(fmt.Sprintf("%T", j), "discarding escaped string in:", string(b[p:end]))
		} else {
			fmt.Println(fmt.Sprintf("%T", j), "consuming escaped string in:", string(b[p:end]))
		}
	}
	for p < end {
		thisChar := b[p]
		if thisChar == ']' {
			return p, ErrUnexpectedListEnd
		}
		if thisChar == '}' {
			return p, ErrUnexpectedMapEnd
		}
		p += 1
		if thisChar == '"' {
			start := p
			for p < end {
				thisChar := b[p]
				p += 1
				if thisChar == '"' {
					if Verbose {
						fmt.Printf("found escaped string in: %#v\n", string(b[start-1:p]))
					}
					if base != 0 {
						*(*string)(unsafe.Pointer(base)) = string(b[start : p-1])
					}
					return p, nil
				}
			}
			return end, ErrNoQuote
		}
	}
	return end, ErrNoQuoteOpen
}

func (j jsonEscapedString) ReportPlan(r *jsonReport) {
	r.Then(`Search for ", returning if I find } or ]`)
	r.Then(`Search for closing "`)
	r.Then(`Create a string in the base from the bytes I found`)
}

type jsonArray struct {
	sliceType    reflect.Type
	internalProc jsonStoredProcedure
	internalType reflect.Type
	cache        byte
}

func newJsonArray(t reflect.Type, d describer) *jsonArray {
	e := t.Elem()
	j := &jsonArray{sliceType: t, internalProc: d.Describe(e)}

	j.internalType = e
	switch e.Kind() {
	case reflect.String:
		j.cache = 's'
	}
	return j
}

func (j jsonArray) ReportPlan(r *jsonReport) {
	r.Then(`Search for [, returning if I find } or ]`)
	r.Then(`Repeatedly...`)
	child := r.Deeper()

	j.internalProc.ReportPlan(r)

	child()
	r.Then(`Write that new array into the pointer`)
}

func (j jsonArray) IntoPointer(op decodeOperation, p, end int, base uintptr) (int, error) {
	if base == 0 && op.mode != ModeSkip {
		panic("bad value here " + fmt.Sprintf("%#v", op.mode))
	}

	b := op.rawData
	if Verbose {
		if op.mode == ModeSkip {
			fmt.Println(fmt.Sprintf("%T", j), "discarding list in", string(b[p:end]))
		} else {
			fmt.Println(fmt.Sprintf("%T", j), "consuming list in", string(b[p:end]))
		}
	}

	store := op.mode == ModeAlloc

	itemSize := int(j.internalType.Size())

	l := 0
	var pointers []byte
	posInPointers := 0

	for p < end {
		thisChar := b[p]
		if thisChar == ']' {
			return p, ErrUnexpectedListEnd
		}
		if thisChar == '}' {
			return p, ErrUnexpectedMapEnd
		}
		p += 1
		if thisChar == '[' {
			start := p - 1
			for p < end {
			anotherItem:
				thisChar := b[p]
				p += 1

				if thisChar == ']' {
					amt := posInPointers / itemSize
					if store {
						switch j.cache {
						case 's':
							if Verbose {
								fmt.Println("we know it's a slice of strings")
							}
							if posInPointers > 0 {
								arr := make([]string, amt, amt)
								s := (*reflect.SliceHeader)(unsafe.Pointer(&arr))
								s.Data = uintptr(unsafe.Pointer(&pointers[0]))
								*(*[]string)(unsafe.Pointer(base)) = arr
							} else {
								arr := make([]string, 0, 0)
								*(*[]string)(unsafe.Pointer(base)) = arr
							}
						default:
							arr := reflect.NewAt(j.sliceType, unsafe.Pointer(base))
							currSlice := reflect.Indirect(arr)
							if posInPointers > 0 {
								asArrayType := reflect.ArrayOf(amt, j.internalType)
								asArray := reflect.NewAt(asArrayType, unsafe.Pointer(&pointers[0]))
								pureArray := reflect.Indirect(asArray)
								nt := pureArray.Slice(0, amt)
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
						}
					}
					if Verbose {
						fmt.Println("found list", string(b[start:p]))
					}
					return p, nil
				}

				var newPtr uintptr
				if store {
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
					newPtr = uintptr(unsafe.Pointer(&pointers[posInPointers-itemSize]))
				}

				if op.mode != ModeSkip && newPtr == 0 {
					panic("bad mode ptr " + fmt.Sprintf(`%#v`, op.mode))
				}
				n, err := j.internalProc.IntoPointer(op, p-1, end, newPtr)
				if err != nil {
					if err != ErrUnexpectedListEnd {
						return n, err
					}
					posInPointers -= itemSize
				}
				p = n
				goto anotherItem
			}
			return end, ErrNoBracket
		}
	}
	return end, ErrNoBracketOpen
}

type jsonMaybeNull struct {
	ptrType           reflect.Type
	underlyingType    reflect.Type
	underlyingHandler jsonStoredProcedure
}

func (j jsonMaybeNull) String() string {
	return fmt.Sprintf(`space-maker for %s`, j.underlyingType.String())
}

func (j jsonMaybeNull) ReportPlan(r *jsonReport) {
	r.Then(`Check to see if I have a nil %s`, j.ptrType)
	func() {
		defer r.Deeper()()
		r.Then(`If so, create a %s`, j.underlyingType)
	}()
	j.underlyingHandler.ReportPlan(r)
}

func newMaybeNull(t reflect.Type, d describer) *jsonMaybeNull {
	underT := t.Elem()
	return &jsonMaybeNull{
		ptrType:           t,
		underlyingType:    underT,
		underlyingHandler: d.Describe(underT),
	}
}

func (j jsonMaybeNull) IntoPointer(op decodeOperation, p, end int, base uintptr) (int, error) {
	if base == 0 && op.mode != ModeSkip {
		panic("bad value here " + fmt.Sprintf("%#v", op.mode))
	}
	b := op.rawData
	// base is a pointer to a potentially-nil pointer to a T
	// we need to ensure it points to an initialized pointer to an initialized T

	if op.mode == ModeSkip {
		return j.underlyingHandler.IntoPointer(op, p, end, 0)
	}
	curPtr := reflect.NewAt(j.ptrType, unsafe.Pointer(base))

	if Verbose {
		fmt.Printf("cutPtr: isnull %t\n", curPtr.IsNil())
		fmt.Printf("%s consuming maybe null at %#v\n", j.String(), curPtr.Interface())
	}

	if curPtr.IsNil() {
		newInstance := reflect.New(j.ptrType)
		curPtr.Set(newInstance)
	}

	if Verbose {
		fmt.Printf("cutPtr: isnull %t\n", curPtr.IsNil())
		fmt.Printf("%s consuming maybe null at %#v\n", j.String(), curPtr.Interface())
	}

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

type jsonInspect struct {
	mapHandler    jsonStoredProcedure
	listHandler   jsonStoredProcedure
	stringHandler jsonStoredProcedure
}

func newJsonInspect() *jsonInspect {
	if Verbose {
		fmt.Println("called newJsonInspect")
	}
	j := &jsonInspect{}
	return j
}

func (j jsonInspect) ReportPlan(r *jsonReport) {
	r.Then(`If I get a {, I'll pass it off as a map[string]interface{}`)
	r.Then(`If I get a [, I'll pass it off as a []interface{}`)
	r.Then(`If I get a ", I'll pass it off as a string`)
	r.Then(`I'll dereference the result into the interface{} in the base pointer`)
}

func (j *jsonInspect) Setup(d describer) {
	j.mapHandler = d.Describe(reflect.TypeOf(make(map[string]interface{})))
	j.listHandler = d.Describe(reflect.TypeOf(make([]interface{}, 0)))
	j.stringHandler = jsonEscapedString{}
}

func (j jsonInspect) IntoPointer(op decodeOperation, p, end int, base uintptr) (int, error) {
	if base == 0 && op.mode != ModeSkip {
		panic("bad value here " + fmt.Sprintf("%#v", op.mode))
	}
	b := op.rawData
	if Verbose {
		if base == 0 {
			fmt.Println(fmt.Sprintf("%T", j), "discarding anything in", string(b[p:end]))
		} else {
			fmt.Println(fmt.Sprintf("%T", j), "consuming anything in", string(b[p:end]))
		}
	}

	asP := (*interface{})(unsafe.Pointer(base))

	for p < end {
		thisChar := b[p]
		if thisChar == ']' {
			return p, ErrUnexpectedListEnd
		}
		if thisChar == '}' {
			return p, ErrUnexpectedMapEnd
		}
		p += 1
		if thisChar == '[' {
			if op.mode == ModeSkip {
				return j.listHandler.IntoPointer(op, p-1, end, 0)
			}
			var l []interface{}
			ptr := unsafe.Pointer(&l)
			n, err := j.listHandler.IntoPointer(op, p-1, end, uintptr(ptr))
			if err != nil {
				return n, err
			}

			if Verbose {
				fmt.Println("inspected array at", string(b[p:n]))
			}

			*asP = l
			return n, nil
		}
		if thisChar == '{' {
			if base == 0 {
				return j.mapHandler.IntoPointer(op, p-1, end, 0)
			}
			var l map[string]interface{}
			ptr := unsafe.Pointer(&l)
			n, err := j.mapHandler.IntoPointer(op, p-1, end, uintptr(ptr))
			if err != nil {
				return n, err
			}

			if Verbose {
				fmt.Println("inspected object at", string(b[p:n]))
			}

			*asP = l
			return n, nil
		}
		if thisChar == '"' {
			if base == 0 {
				return j.stringHandler.IntoPointer(op, p-1, end, 0)
			}
			var l string
			ptr := unsafe.Pointer(&l)
			n, err := j.stringHandler.IntoPointer(op, p-1, end, uintptr(ptr))
			if err != nil {
				return n, err
			}

			if Verbose {
				fmt.Println("inspected escaped string at", string(b[p:n]))
			}

			*asP = l
			return n, nil
		}
	}
	return 0, ErrUnexpectedEOF
}

type jsonStringMap struct{}

func (j jsonStringMap) ReportPlan(r *jsonReport) {
	r.Then(`Look for a {, create a map[string]string, then repeatedly:`)
	func() {
		defer r.Deeper()()

		func() {
			r.Then("To get a key, I'll:")
			defer r.Deeper()()
			jsonRawString{}.ReportPlan(r)
		}()
		func() {
			r.Then("To get a value, I'll:")
			defer r.Deeper()()
			jsonEscapedString{}.ReportPlan(r)
		}()
		r.Then("Save it in the map")
	}()
	r.Then(`Store the new map in the base pointer`)
}

func (j jsonStringMap) IntoPointer(op decodeOperation, p, end int, base uintptr) (int, error) {
	if base == 0 && op.mode != ModeSkip {
		panic("bad value here " + fmt.Sprintf("%#v", op.mode))
	}
	b := op.rawData
	if Verbose {
		if op.mode == ModeSkip {
			fmt.Println(fmt.Sprintf("%T", j), "discarding stringmap in", string(b[p:end]))
		} else {
			fmt.Println(fmt.Sprintf("%T", j), "consuming stringmap in", string(b[p:end]))
		}
	}

	store := op.mode == ModeAlloc

	var cMap map[string]string
	var lStr, rStr string
	var lPtr, rPtr uintptr
	if store {
		currentMap := (*map[string]string)(unsafe.Pointer(base))
		if *currentMap == nil {
			*currentMap = make(map[string]string, DefaultMapSize)
		}
		cMap = *currentMap
		lPtr, rPtr = uintptr(unsafe.Pointer(&lStr)), uintptr(unsafe.Pointer(&rStr))
	}

	for p < end {
		thisChar := b[p]

		if thisChar == ']' {
			return p, ErrUnexpectedListEnd
		}
		if thisChar == '}' {
			return p, ErrUnexpectedMapEnd
		}

		p += 1
		if thisChar == '{' {
			mapStart := p - 1
			for p < end {
			anotherKey:
				n, err := jsonRawString{}.IntoPointer(op, p, end, lPtr)
				if err != nil {
					if err != ErrUnexpectedMapEnd {
						return n, err
					}
					if Verbose {
						fmt.Println("found map:", string(b[mapStart:n+1]))
					}
					return n + 1, nil
				}

				p = n
				for p < end {
					thisChar := b[p]
					p += 1
					if thisChar == ':' {
						n, err := jsonEscapedString{}.IntoPointer(op, p, end, rPtr)
						if err != nil {
							return n, err
						}
						p = n
						if store {
							cMap[lStr] = rStr
						}
						goto anotherKey
					}
				}
				return end, ErrNoQuote
			}
			return end, ErrNoBrace
		}
	}
	return end, ErrNoBrace
}

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

func (j jsonMap) ReportPlan(r *jsonReport) {
	r.Then(`Look for a {, create a ` + j.all.String() + ` in the base pointer using reflection, then repeatedly:`)
	func() {
		defer r.Deeper()()
		func() {
			r.Then("To get a key, I'll:")
			defer r.Deeper()()
			j.left.ReportPlan(r)
		}()
		func() {
			r.Then("To get a value, I'll:")
			defer r.Deeper()()
			j.right.ReportPlan(r)
		}()
		r.Then("Save the key and value into the map using reflection")
	}()
}

func (j jsonMap) IntoPointer(op decodeOperation, p, end int, base uintptr) (int, error) {
	if base == 0 && op.mode != ModeSkip {
		panic("bad value here " + fmt.Sprintf("%#v", op.mode))
	}
	b := op.rawData
	if Verbose {
		if op.mode == ModeSkip {
			fmt.Println(fmt.Sprintf("%T", j), "discarding map in", string(b[p:end]))
		} else {
			fmt.Println(fmt.Sprintf("%T", j), "consuming map in", string(b[p:end]))
		}
	}

	store := op.mode == ModeAlloc
	var lSide, rSide, currentMap reflect.Value
	var lPtr, rPtr uintptr

	if store {
		currentMap = reflect.Indirect(reflect.NewAt(j.all, unsafe.Pointer(base)))
		if reflect.Indirect(currentMap).IsNil() {
			newMap := reflect.Indirect(reflect.MakeMapWithSize(j.all, DefaultMapSize))
			currentMap.Set(newMap)
		}

		// itemSize := j.rightType.Size()

		lSide = reflect.New(j.leftType)
		rSide = reflect.New(j.rightType)

		lPtr = lSide.Pointer()
		rPtr = rSide.Pointer()
	} else {
		lPtr = 0
		rPtr = 0
	}

	for p < end {
		thisChar := b[p]

		if thisChar == ']' {
			return p, ErrUnexpectedListEnd
		}
		if thisChar == '}' {
			return p, ErrUnexpectedMapEnd
		}

		p += 1
		if thisChar == '{' {
			mapStart := p - 1
			for p < end {
			anotherKey:
				n, err := j.left.IntoPointer(op, p, end, lPtr)
				if err != nil {
					if err != ErrUnexpectedMapEnd {
						return n, err
					}
					if Verbose {
						fmt.Println("found map:", string(b[mapStart:n+1]))
					}
					return n + 1, nil
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
						if store {
							currentMap.SetMapIndex(reflect.Indirect(lSide), reflect.Indirect(rSide))
						}
						goto anotherKey
					}
				}
				return end, ErrNoColon
			}
		}
	}
	return end, ErrNoBrace
}

type field struct {
	natural bool
	offset  uintptr
	bytes   []byte
}

func (f field) String() string {
	return fmt.Sprintf(`%s at %d`, string(f.bytes), f.offset)
}

func (f field) Less(than []byte) bool {
	i := 0
	for {
		if i == len(f.bytes) {
			return true
		}
		if i == len(than) {
			return false
		}

		if f.bytes[i] != than[i] {
			return f.bytes[i] < than[i]
		}

		i += 1
	}
}

func (f field) Greater(than []byte) bool {
	i := 0
	for {
		if i == len(than) {
			return true
		}
		if i == len(f.bytes) {
			return false
		}

		if f.bytes[i] != than[i] {
			return f.bytes[i] > than[i]
		}

		i += 1
	}
}

func (f field) Equal(than []byte) bool {
	i := 0
	if len(f.bytes) != len(than) {
		return false
	}

	for {
		if i == len(f.bytes) {
			return true
		}
		if f.bytes[i] != than[i] {
			return f.bytes[i] < than[i]
		}
		i += 1
	}
}

type fields []field

func (f fields) Len() int {
	return len(f)
}

func (f fields) Less(a, b int) (res bool) {
	return f[a].Less(f[b].bytes)
}

func (f fields) Swap(a, b int) {
	tmp := f[b]
	f[b] = f[a]
	f[a] = tmp
}

type jsonObject struct {
	fields     fields
	offsets    []jsonStoredProcedure
	def        jsonStoredProcedure
	structType reflect.Type
}

func (j *jsonObject) addName(name string, offset uintptr, natural bool) {
	j.fields = append(j.fields, field{offset: offset, bytes: []byte(name), natural: natural})
}

func (j jsonObject) String() string {
	return fmt.Sprintf("json object mapping to %s", j.structType.String())
}

func newJsonObject(obj reflect.Type, des describer) *jsonObject {
	offsets := make([]jsonStoredProcedure, int(obj.Size()), int(obj.Size()))
	j := &jsonObject{offsets: offsets}
	for i := 0; i < obj.NumField(); i++ {
		f := obj.Field(i)
		j.addName(f.Name, f.Offset, true)
		j.addName(strings.ToLower(f.Name), f.Offset, false)
		j.addName(strings.ToUpper(f.Name), f.Offset, false)
		offsets[f.Offset] = des.Describe(f.Type)
		if Verbose {
			fmt.Printf("jsonObject: for %s.%s, use %#v\n", obj.String(), f.Name, offsets[f.Offset])
		}
	}

	j.structType = obj

	var anything []interface{}
	j.def = des.Describe(reflect.TypeOf(anything).Elem())

	if Verbose {
		fmt.Println("sorting fields", j.fields)
	}
	sort.Sort(j.fields)

	if Verbose {
		fmt.Println(obj.String(), j.fields)
	}
	return j
}

func (j jsonObject) ReportPlan(r *jsonReport) {
	r.Then(`Look for a {, then repeatedly:`)
	func() {
		defer r.Deeper()()
		r.Then("Get a key by scanning for raw bytes")
		r.Then("Binary search for that key through %d handlers", len(j.fields))

		for _, f := range j.fields {
			if !f.natural {
				continue
			}
			r.Then("If the key is like %s, I'll:", string(f.bytes))
			func() {
				defer r.Deeper()()
				j.offsets[f.offset].ReportPlan(r)
			}()
		}
		r.Then("If it's any other key, I'll:")
		func() {
			defer r.Deeper()()
			j.def.ReportPlan(r)
		}()
	}()
}

func (j jsonObject) IntoPointer(op decodeOperation, p, end int, base uintptr) (int, error) {
	if base == 0 && op.mode != ModeSkip {
		panic("bad value here " + fmt.Sprintf("%#v", op.mode))
	}
	b := op.rawData
	if Verbose {
		fmt.Println(j.String(), "consuming object in", string(b[p:end]))
	}

	for p < end {
		thisChar := b[p]

		if thisChar == ']' {
			return p, ErrUnexpectedListEnd
		}
		if thisChar == '}' {
			return p, ErrUnexpectedMapEnd
		}

		p += 1
		if thisChar == '{' {
			var handler jsonStoredProcedure
			var offset uintptr

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
							handler = j.def
							offset = 0

							if Verbose {
								fmt.Println("found key", string(b[start:p-1]))
							}

							bytes := b[start : p-1]

							fs := j.fields
							var foundN int

							{ // find an N of a field using binary search
								i, j := 0, len(fs)
								for i < j {
									h := (i + j) >> 1
									// i ≤ h < j
									if !fs[h].Greater(bytes) {
										i = h + 1
									} else {
										j = h
									}
								}
								foundN = i
							}

							if Verbose {
								fmt.Println("searching for key returned", foundN, "/", len(j.fields))
							}

							if foundN < len(j.fields) {
								f := j.fields[foundN]
								if f.Equal(bytes) {
									if Verbose {
										fmt.Println("found handler for key", f)
									}
									offset = base + f.offset
									handler = j.offsets[f.offset]
								} else {
									if Verbose {
										fmt.Println("key was not found")
									}
								}
							} else if Verbose {
								fmt.Println("key was not found")
							}

							for p < end {
								thisChar := b[p]
								p += 1
								if thisChar == ':' {
									op := op
									if offset == 0 {
										op.mode = ModeSkip
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
	return end, ErrNoBraceOpen
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
					ids = append(ids, [3]int{start, strEnd, 0})
					goto anotherElement
				}
			}
		}
	}
	return
}

type ParsingMode byte

const (
	ModeAlloc ParsingMode = 'u'
	ModeSkip  ParsingMode = 's'
)

type decodeOperation struct {
	rawData []byte
	mode    ParsingMode
	done    chan bool
	desc    jsonStoredProcedure
}

type Describers struct {
	allTypes     sync.Map
	pendingTypes sync.Map
	generic      jsonStoredProcedure
	lookAheads   chan decodeOperation
}

func NewDescriber() *Describers {
	d := &Describers{lookAheads: make(chan decodeOperation, runtime.NumCPU())}
	d.Store(reflect.TypeOf(map[string]string{}), jsonStringMap{})

	for i := runtime.NumCPU(); i > 0; i-- {
		go func() {
			for {
				op := <-d.lookAheads
				op.mode = ModeSkip
				op.desc.IntoPointer(op, 0, len(op.rawData), 0)
				close(op.done)
			}
		}()
	}

	in := newJsonInspect()
	var i []interface{}
	d.Store(reflect.TypeOf(i).Elem(), in)
	in.Setup(d)

	return d
}

func (d *Describers) LearnAbout(t reflect.Type) jsonStoredProcedure {
	if t == nil {
		panic("can't learn about nil type")
	}
	if Verbose {
		fmt.Println("learning about", t.String())
	}
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

func (d *Describers) Store(t reflect.Type, proc jsonStoredProcedure) {
	if Verbose {
		fmt.Printf("encoder for the %d byte %s = %T\n", t.Size(), t.String(), proc)
	}
	d.allTypes.Store(t, proc)
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
	d.Store(t, newProc)

	lockIt.Broadcast()
	d.pendingTypes.Delete(t)

	return newProc
}

func (d *Describers) ReportPlan(sample interface{}) jsonReport {
	j := &jsonReport{}
	j.Then("Here's how I plan to decode %T", sample)
	des := d.Describe(reflect.TypeOf(sample))
	des.ReportPlan(j)
	return *j
}

func (d *Describers) Unmarshal(b []byte, to interface{}) error {
	t := reflect.TypeOf(to)
	desc := d.Describe(t)

	op := decodeOperation{desc: desc, rawData: b, mode: ModeAlloc}
	if LookAhead {
		op.done = make(chan bool)
		d.lookAheads <- op
	}

	if !DryRun {
		// create a pointer to whatever i've been given
		v := reflect.ValueOf(to)

		newPtr := reflect.NewAt(t, unsafe.Pointer(v.Pointer()))

		if Verbose {
			fmt.Println("created space", newPtr.String(), reflect.Indirect(newPtr).Interface(), newPtr.Pointer())
			fmt.Printf("calling out to %T\n", desc)
		}

		reflect.Indirect(newPtr).Set(v)

		_, err := desc.IntoPointer(op, 0, len(b), newPtr.Pointer())
		if err != nil {
			return err
		}
	}

	if LookAhead {
		<-op.done
	}
	return nil
}

var Standard = NewDescriber()

func Unmarshal(b []byte, to interface{}) error {
	return Standard.Unmarshal(b, to)
}
