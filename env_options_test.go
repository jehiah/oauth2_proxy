package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type envTest struct {
	testField string `cfg:"target_field" env:"TEST_ENV_FIELD"`
}

func TestLoadEnvForStruct(t *testing.T) {
	var ensureUsage envTest
	_ = ensureUsage.testField

	cfg := make(EnvOptions)
	cfg.LoadEnvForStruct(&envTest{})

	_, ok := cfg["target_field"]
	assert.Equal(t, ok, false)

	if err := os.Setenv("TEST_ENV_FIELD", "1234abcd"); err != nil {
		t.Fatalf("failed to set env: %v", err)
	}
	cfg.LoadEnvForStruct(&envTest{})
	v := cfg["target_field"]
	assert.Equal(t, v, "1234abcd")
}
