package utils_test

import (
	"fmt"
	"testing"

	"gitea.chrnv.ru/vanadium23/kompanion/pkg/utils"
)

func TestPartialMd5(t *testing.T) {
	expected := "5ee88058c4346a122c4ccf80e36b1dc8"
	actual, err := utils.PartialMD5("../../test/test_data/CrimePunishment-EPUB2.epub")
	if err != nil {
		t.Fatalf("Error calculating MD5: %v", err)
	}
	if expected != fmt.Sprintf("%x", actual) {
		t.Fatalf("Expected MD5 %s, got %x", expected, actual)
	}
}
