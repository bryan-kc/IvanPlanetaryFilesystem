package server

import (
	"proj2_f5w9a_h6v9a_q7w9a_r8u8_w1c0b/serverpb"
	"reflect"
	"testing"
)

func filter(t *testing.T, msg ...string) *serverpb.BloomFilter {
	filter := createNewBloomFilter()
	for _, m := range msg {
		filter.AddString(m)
	}

	mergedFilter, err := filter.GobEncode()
	if err != nil {
		t.Fatal(err)
	}

	return &serverpb.BloomFilter{Data: mergedFilter}
}

func TestDeleteDuplicates(t *testing.T) {
	empty := &serverpb.BloomFilter{}
	emptyString := filter(t)
	a := filter(t, "a")
	b := filter(t, "b")
	ab := filter(t, "a", "b")

	cases := []struct {
		in, want []*serverpb.BloomFilter
		maxWidth int
	}{
		{
			[]*serverpb.BloomFilter{
				empty,
			},
			[]*serverpb.BloomFilter{
				empty,
			},
			0,
		},
		{
			[]*serverpb.BloomFilter{
				a, b, ab,
			},
			[]*serverpb.BloomFilter{
				a, b, ab,
			},
			0,
		},
		{
			[]*serverpb.BloomFilter{
				empty, empty, a, b, ab, a, b,
			},
			[]*serverpb.BloomFilter{
				empty, empty, a, b, ab,
			},
			0,
		},
		{
			[]*serverpb.BloomFilter{
				empty, empty, emptyString, emptyString, a, empty, emptyString,
			},
			[]*serverpb.BloomFilter{
				empty, empty, emptyString, emptyString, a,
			},
			0,
		},
		// Max Width Tests
		{
			[]*serverpb.BloomFilter{
				a, b, ab,
			},
			[]*serverpb.BloomFilter{
				a, b, ab,
			},
			10,
		},
		{
			[]*serverpb.BloomFilter{
				a, b, ab,
			},
			[]*serverpb.BloomFilter{
				a, b,
			},
			2,
		},
	}

	for i, c := range cases {
		out, err := deleteDuplicates(c.in, c.maxWidth)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(out, c.want) {
			t.Errorf("%d. got %+v, want %+v", i, len(out), len(c.want))
		}
	}
}
