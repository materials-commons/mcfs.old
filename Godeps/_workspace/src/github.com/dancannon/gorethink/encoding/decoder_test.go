package encoding

import (
	"image"
	"reflect"
	"testing"
)

type T struct {
	X string
	Y int
	Z int `gorethink:"-"`
}

type U struct {
	Alphabet string `gorethink:"alpha"`
}

type V struct {
	F1 interface{}
	F2 int32
	F3 string
}

type tx struct {
	x int
}

var txType = reflect.TypeOf((*tx)(nil)).Elem()

// Test data structures for anonymous fields.

type Point struct {
	Z int
}

type Top struct {
	Level0 int
	Embed0
	*Embed0a
	*Embed0b `gorethink:"e,omitempty"` // treated as named
	Embed0c  `gorethink:"-"`           // ignored
	Loop
	Embed0p // has Point with X, Y, used
	Embed0q // has Point with Z, used
}

type Embed0 struct {
	Level1a int // overridden by Embed0a's Level1a with tag
	Level1b int // used because Embed0a's Level1b is renamed
	Level1c int // used because Embed0a's Level1c is ignored
	Level1d int // annihilated by Embed0a's Level1d
	Level1e int `gorethink:"x"` // annihilated by Embed0a.Level1e
}

type Embed0a struct {
	Level1a int `gorethink:"Level1a,omitempty"`
	Level1b int `gorethink:"LEVEL1B,omitempty"`
	Level1c int `gorethink:"-"`
	Level1d int // annihilated by Embed0's Level1d
	Level1f int `gorethink:"x"` // annihilated by Embed0's Level1e
}

type Embed0b Embed0

type Embed0c Embed0

type Embed0p struct {
	image.Point
}

type Embed0q struct {
	Point
}

type Loop struct {
	Loop1 int `gorethink:",omitempty"`
	Loop2 int `gorethink:",omitempty"`
	*Loop
}

// From reflect test:
// The X in S6 and S7 annihilate, but they also block the X in S8.S9.
type S5 struct {
	S6
	S7
	S8
}

type S6 struct {
	X int
}

type S7 S6

type S8 struct {
	S9
}

type S9 struct {
	X int
	Y int
}

// From reflect test:
// The X in S11.S6 and S12.S6 annihilate, but they also block the X in S13.S8.S9.
type S10 struct {
	S11
	S12
	S13
}

type S11 struct {
	S6
}

type S12 struct {
	S6
}

type S13 struct {
	S8
}

type decodeTest struct {
	in  interface{}
	ptr interface{}
	out interface{}
	err error
}

type Ambig struct {
	// Given "hello", the first match should win.
	First  int `gorethink:"HELLO"`
	Second int `gorethink:"Hello"`
}

var decodeTests = []decodeTest{
	// basic types
	{in: true, ptr: new(bool), out: true},
	{in: 1, ptr: new(int), out: 1},
	{in: 1.2, ptr: new(float64), out: 1.2},
	{in: -5, ptr: new(int16), out: int16(-5)},
	{in: 2, ptr: new(string), out: string("2")},
	{in: float64(2.0), ptr: new(interface{}), out: float64(2.0)},
	{in: string("2"), ptr: new(interface{}), out: string("2")},
	{in: "a\u1234", ptr: new(string), out: "a\u1234"},
	{in: map[string]interface{}{"X": []interface{}{1, 2, 3}, "Y": 4}, ptr: new(T), out: T{}, err: &DecodeTypeError{"array", reflect.TypeOf("")}},
	{in: map[string]interface{}{"x": 1}, ptr: new(tx), out: tx{}},
	{in: map[string]interface{}{"F1": float64(1), "F2": 2, "F3": 3}, ptr: new(V), out: V{F1: float64(1), F2: int32(2), F3: string("3")}},
	{in: map[string]interface{}{"F1": string("1"), "F2": 2, "F3": 3}, ptr: new(V), out: V{F1: string("1"), F2: int32(2), F3: string("3")}},
	{
		in:  map[string]interface{}{"k1": int64(1), "k2": "s", "k3": []interface{}{int64(1), 2.0, 3e-3}, "k4": map[string]interface{}{"kk1": "s", "kk2": int64(2)}},
		out: map[string]interface{}{"k1": int64(1), "k2": "s", "k3": []interface{}{int64(1), 2.0, 3e-3}, "k4": map[string]interface{}{"kk1": "s", "kk2": int64(2)}},
		ptr: new(interface{}),
	},

	// Z has a "-" tag.
	{in: map[string]interface{}{"Y": 1, "Z": 2}, ptr: new(T), out: T{Y: 1}},

	{in: map[string]interface{}{"alpha": "abc", "alphabet": "xyz"}, ptr: new(U), out: U{Alphabet: "abc"}},
	{in: map[string]interface{}{"alpha": "abc"}, ptr: new(U), out: U{Alphabet: "abc"}},
	{in: map[string]interface{}{"alphabet": "xyz"}, ptr: new(U), out: U{}},

	// array tests
	{in: []interface{}{1, 2, 3}, ptr: new([3]int), out: [3]int{1, 2, 3}},
	{in: []interface{}{1, 2, 3}, ptr: new([1]int), out: [1]int{1}},
	{in: []interface{}{1, 2, 3}, ptr: new([5]int), out: [5]int{1, 2, 3, 0, 0}},

	// empty array to interface test
	{in: []interface{}{}, ptr: new([]interface{}), out: []interface{}{}},
	{in: map[string]interface{}{"T": []interface{}{}}, ptr: new(map[string]interface{}), out: map[string]interface{}{"T": []interface{}{}}},

	{
		in: map[string]interface{}{
			"Level0":  1,
			"Level1b": 2,
			"Level1c": 3,
			"level1d": 4,
			"Level1a": 5,
			"LEVEL1B": 6,
			"e": map[string]interface{}{
				"Level1a": 8,
				"Level1b": 9,
				"Level1c": 10,
				"Level1d": 11,
				"x":       12,
			},
			"Loop1": 13,
			"Loop2": 14,
			"X":     15,
			"Y":     16,
			"Z":     17,
		},
		ptr: new(Top),
		out: Top{
			Level0: 1,
			Embed0: Embed0{
				Level1b: 2,
				Level1c: 3,
			},
			Embed0a: &Embed0a{
				Level1a: 5,
				Level1b: 6,
			},
			Embed0b: &Embed0b{
				Level1a: 8,
				Level1b: 9,
				Level1c: 10,
				Level1d: 11,
				Level1e: 12,
			},
			Loop: Loop{
				Loop1: 13,
				Loop2: 14,
			},
			Embed0p: Embed0p{
				Point: image.Point{X: 15, Y: 16},
			},
			Embed0q: Embed0q{
				Point: Point{Z: 17},
			},
		},
	},
	{
		in:  map[string]interface{}{"hello": 1},
		ptr: new(Ambig),
		out: Ambig{First: 1},
	},

	{
		in:  map[string]interface{}{"X": 1, "Y": 2},
		ptr: new(S5),
		out: S5{S8: S8{S9: S9{Y: 2}}},
	},
	{
		in:  map[string]interface{}{"X": 1, "Y": 2},
		ptr: new(S10),
		out: S10{S13: S13{S8: S8{S9: S9{Y: 2}}}},
	},
}

func TestDecode(t *testing.T) {
	for i, tt := range decodeTests {
		if tt.ptr == nil {
			continue
		}

		// v = new(right-type)
		v := reflect.New(reflect.TypeOf(tt.ptr).Elem())

		err := Decode(v.Interface(), tt.in)
		if tt.err != nil {
			if !reflect.DeepEqual(err, tt.err) {
				t.Errorf("#%d: got error %v want %v", i, err, tt.err)
			}

			continue
		}

		if !reflect.DeepEqual(v.Elem().Interface(), tt.out) {
			t.Errorf("#%d: mismatch\nhave: %+v\nwant: %+v", i, v.Elem().Interface(), tt.out)
			continue
		}

		// Check round trip.
		if tt.err == nil {
			enc, err := Encode(v.Interface())
			if err != nil {
				t.Errorf("#%d: error re-marshaling: %v", i, err)
				continue
			}
			vv := reflect.New(reflect.TypeOf(tt.ptr).Elem())

			if err := Decode(vv.Interface(), enc); err != nil {
				t.Errorf("#%d: error re-decodeing: %v", i, err)
				continue
			}
			if !reflect.DeepEqual(v.Elem().Interface(), vv.Elem().Interface()) {
				t.Errorf("#%d: mismatch\nhave: %#+v\nwant: %#+v", i, v.Elem().Interface(), vv.Elem().Interface())
				continue
			}
		}
	}
}

func TestStringKind(t *testing.T) {
	type aMap map[string]int

	var m1, m2 map[string]int
	m1 = map[string]int{
		"foo": 42,
	}

	data, err := Encode(m1)
	if err != nil {
		t.Errorf("Unexpected error encoding: %v", err)
	}

	err = Decode(&m2, data)
	if err != nil {
		t.Errorf("Unexpected error decoding: %v", err)
	}

	if !reflect.DeepEqual(m1, m2) {
		t.Error("Items should be equal after encoding and then decoding")
	}

}

var decodeTypeErrorTests = []struct {
	dest interface{}
	src  interface{}
}{
	{new(string), map[interface{}]interface{}{"user": "name"}},
	{new(error), map[interface{}]interface{}{}},
	{new(error), []interface{}{}},
	{new(error), ""},
	{new(error), 123},
	{new(error), true},
}

func TestDecodeTypeError(t *testing.T) {
	for _, item := range decodeTypeErrorTests {
		err := Decode(item.dest, item.src)
		if _, ok := err.(*DecodeTypeError); !ok {
			t.Errorf("expected type error for Decode(%q, type %T): got %T",
				item.src, item.dest, err)
		}
	}
}

// Test handling of unexported fields that should be ignored.
type unexportedFields struct {
	Name string
	m    map[string]interface{} `gorethink:"-"`
	m2   map[string]interface{} `gorethink:"abcd"`
}

func TestDecodeUnexported(t *testing.T) {
	input := map[string]interface{}{
		"Name": "Bob",
		"m": map[string]interface{}{
			"x": 123,
		},
		"m2": map[string]interface{}{
			"y": 123,
		},
		"abcd": map[string]interface{}{
			"z": 789,
		},
	}
	want := &unexportedFields{Name: "Bob"}

	out := &unexportedFields{}
	err := Decode(out, input)
	if err != nil {
		t.Errorf("got error %v, expected nil", err)
	}
	if !reflect.DeepEqual(out, want) {
		t.Errorf("got %q, want %q", out, want)
	}
}

type Foo struct {
	FooBar interface{} `gorethink:"foobar"`
}
type Bar struct {
	Baz int `gorethink:"baz"`
}
