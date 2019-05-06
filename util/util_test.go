package util_test

import (
	"testing"

	"github.com/magneticio/vampkubistcli/util"
	"github.com/stretchr/testify/assert"
)

func TestMergeBasic(t *testing.T) {
	destination :=
		`metadata:
  field1: value1
  field2: value2
  `
	source :=
		`metadata:
  field1: value1updated
  field3: value3
  `

	expectedMerged :=
		`metadata:
  field1: value1updated
  field2: value2
  field3: value3
`
	serializationType := "yaml"
	merged, err := util.Merge(destination, source, serializationType)

	assert.NoError(t, err)
	assert.Equal(t, expectedMerged, merged)
}

func TestMergeMapInMap(t *testing.T) {
	destination :=
		`metadata:
  field1:
    value1map:
      key1: innerValue1
      key2: 1
      key3: noUpdate
  field2: value2
  `
	source :=
		`metadata:
  field1:
    value1map:
      key1: innerValue1updated
      key2: 2
  field3: value3
  `

	expectedMerged :=
		`metadata:
  field1:
    value1map:
      key1: innerValue1updated
      key2: 2
      key3: noUpdate
  field2: value2
  field3: value3
`
	serializationType := "yaml"
	merged, err := util.Merge(destination, source, serializationType)

	assert.NoError(t, err)
	assert.Equal(t, expectedMerged, merged)
}
